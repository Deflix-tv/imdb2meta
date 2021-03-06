package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v2"
	"go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"

	"github.com/deflix-tv/imdb2meta/pb"
)

var (
	tsvPath = flag.String("tsvPath", "", `Path to the "data.tsv" file that's inside the "title.basics.tsv.gz" archive`)

	badgerPath = flag.String("badgerPath", "", "Path to the directory with the BadgerDB files")
	boltPath   = flag.String("boltPath", "", "Path to the bbolt DB file")

	limit = flag.Int("limit", 0, "Limit the number of rows to process (excluding the header row)")

	skipEpisodes = flag.Bool("skipEpisodes", false, "Skip storing individual TV episodes")
	skipMisc     = flag.Bool("skipMisc", false, `Skip title types like "videoGame", "audiobook" and "radioSeries"`)
	minimal      = flag.Bool("minimal", false, "Only store minimal metadata (ID, type, title, release/start year)")
)

var (
	imdbBytes       = []byte("imdb") // Bucket name for bbolt
	expectedColumns = 9
)

func main() {
	// Workaround for exiting with 1 despite not using log.Fatal while still running deferred DB close calls.
	exitCode := 1
	defer func() {
		os.Exit(exitCode)
	}()

	flag.Parse()

	// CLI argument check
	if *tsvPath == "" {
		log.Fatalln(`Missing CLI argument "-tsvPath"`)
	}
	if *badgerPath == "" && *boltPath == "" {
		log.Fatalln(`Missing an argument for the DB: Either "-badgerPath" or "-boltPath".`)
	} else if *badgerPath != "" && *boltPath != "" {
		log.Fatalln(`You can only use either "-badgerPath" or "-boltPath", but not both at the same time`)
	}

	f, err := os.Open(*tsvPath)
	if err != nil {
		log.Fatalf("Couldn't open TSV file: %v\n", err)
	}

	s := bufio.NewScanner(f)

	i := 1
	// The first row is just the headers
	if !s.Scan() || len(strings.Split(s.Text(), "\t")) != expectedColumns {
		log.Fatalf("The TSV file doesn't seem to contain any data: %v\n", err)
	}

	var badgerDB *badger.DB
	var boltDB *bbolt.DB
	if *badgerPath != "" {
		opts := badger.DefaultOptions(*badgerPath).
			WithLoggingLevel(badger.WARNING).
			WithSyncWrites(false)
		badgerDB, err = badger.Open(opts)
		if err != nil {
			log.Fatalf("Couldn't open BadgerDB: %v\n", err)
		}
		defer badgerDB.Close()
	} else {
		boltDB, err = bbolt.Open(*boltPath, 0666, nil)
		if err != nil {
			log.Fatalf("Couldn't open bbolt DB: %v\n", err)
		}
		defer boltDB.Close()
		err = boltDB.Update(func(tx *bbolt.Tx) error {
			if tx.Bucket(imdbBytes) == nil {
				_, err := tx.CreateBucket(imdbBytes)
				return err
			}
			return nil
		})
		if err != nil {
			log.Fatalf("Couldn't create bucket in bbolt: %v\n", err)
		}
	}

	// Here after we have opened the DB, don't use log.Fatal or os.Exit, as then the DB won't be closed and can end up in a corrupted state.
	// So we log with Print and then return, leading to the deferred DB close and then deferred os.Exit(1) being called.

	storedCount := 0
	start := time.Now()
	for ; *limit == 0 || i <= *limit; i++ {
		if !s.Scan() {
			break
		}
		if err != nil {
			log.Printf("Couldn't read TSV row %v: %v\n", i, err)
			return
		}

		record := strings.Split(s.Text(), "\t")
		if len(record) != expectedColumns {
			log.Printf("The row didn't have the expected number of columns (row %v): %#v\n", i, record)
			return
		}
		m, err := toMeta(record, *minimal)
		if err != nil {
			log.Printf("Couldn't create Meta from record at row %v: %#v: %v\n", i, record, err)
			return
		}

		// Skip all episodes if configured
		if *skipEpisodes &&
			(m.GetTitleType() == pb.TitleType_TV_EPISODE ||
				m.GetTitleType() == pb.TitleType_EPISODE) {
			continue
		}
		// Skip other stuff if configured
		if *skipMisc &&
			(m.GetTitleType() == pb.TitleType_VIDEO_GAME ||
				m.GetTitleType() == pb.TitleType_AUDIOBOOK ||
				m.GetTitleType() == pb.TitleType_RADIO_SERIES) {
			continue
		}

		mBytes, err := proto.Marshal(m)
		if err != nil {
			log.Printf("Couldn't marshal Meta to protocol buffer at row %v: %+v: %v\n", i, m, err)
			return
		}

		if *badgerPath != "" {
			requiresUpdate := false
			_ = badgerDB.View(func(txn *badger.Txn) error {
				// err can be badger.ErrKeyNotFound and other errors. In any case we want to write to the DB.
				item, err := txn.Get([]byte(m.GetId()))
				if err != nil {
					requiresUpdate = true
					return nil
				}
				// Also write to the DB if the values differ
				_ = item.Value(func(val []byte) error {
					if !bytes.Equal(val, mBytes) {
						requiresUpdate = true
					}
					return nil
				})
				return nil
			})
			if requiresUpdate {
				err = badgerDB.Update(func(txn *badger.Txn) error {
					storedCount++
					return txn.Set([]byte(m.GetId()), mBytes)
				})
			}
		} else {
			requiresUpdate := false
			_ = boltDB.View(func(tx *bbolt.Tx) error {
				txBytes := tx.Bucket(imdbBytes).Get([]byte(m.GetId()))
				if txBytes == nil {
					requiresUpdate = true
					return nil
				}
				// Also write to the DB if the values differ
				if !bytes.Equal(txBytes, mBytes) {
					requiresUpdate = true
				}
				return nil
			})
			if requiresUpdate {
				err = boltDB.Update(func(tx *bbolt.Tx) error {
					storedCount++
					return tx.Bucket(imdbBytes).Put([]byte(m.GetId()), mBytes)
				})
			}
		}
		if err != nil {
			log.Printf("Couldn't write marshalled Meta to database at row %v: %+v: %v\n", i, m, err)
			return
		}

		// Including the header, we've processed i+1 at this point, but it's only going to be incremented at the beginning of the next iteration.
		if (i+1)%1000 == 0 {
			log.Printf("Processed %v rows, stored %v objects\n", i+1, storedCount)
		}
	}
	end := time.Now()
	log.Printf("Processing finished. Processed %v rows, stored %v objects.\n", i, storedCount)
	log.Printf("Processing took %v\n", end.Sub(start))
	exitCode = 0
}

// toMeta converts a TSV record into a Meta object.
func toMeta(record []string, minimal bool) (*pb.Meta, error) {
	meta := &pb.Meta{}

	meta.Id = record[0]

	// As of 2021-01-15 the following title types exist:
	switch record[1] {
	case "movie":
		meta.TitleType = pb.TitleType_MOVIE
	case "short":
		meta.TitleType = pb.TitleType_SHORT
	case "tvEpisode":
		meta.TitleType = pb.TitleType_TV_EPISODE
	case "tvMiniSeries":
		meta.TitleType = pb.TitleType_TV_MINI_SERIES
	case "tvMovie":
		meta.TitleType = pb.TitleType_TV_MOVIE
	case "tvSeries":
		meta.TitleType = pb.TitleType_TV_SERIES
	case "tvShort":
		meta.TitleType = pb.TitleType_TV_SHORT
	case "tvSpecial":
		meta.TitleType = pb.TitleType_TV_SPECIAL
	case "video":
		meta.TitleType = pb.TitleType_VIDEO
	case "videoGame":
		meta.TitleType = pb.TitleType_VIDEO_GAME
	case "audiobook":
		meta.TitleType = pb.TitleType_AUDIOBOOK
	case "radioSeries":
		meta.TitleType = pb.TitleType_RADIO_SERIES
	case "episode":
		meta.TitleType = pb.TitleType_EPISODE
	default:
		return nil, fmt.Errorf("Unknown title type: %v", record[1])
	}

	meta.PrimaryTitle = record[2]

	if record[3] != meta.PrimaryTitle {
		meta.OriginalTitle = record[3]
	}

	if !minimal && record[4] == "1" {
		meta.IsAdult = true
	}

	if record[5] != "\\N" {
		startYear, err := strconv.Atoi(record[5])
		if err != nil {
			return nil, fmt.Errorf("couldn't convert string to int for startYear: %v", err)
		}
		meta.StartYear = int32(startYear)
	}

	if !minimal && record[6] != "\\N" {
		endYear, err := strconv.Atoi(record[6])
		if err != nil {
			return nil, fmt.Errorf("couldn't convert string to int for endYear: %v", err)
		}
		meta.EndYear = int32(endYear)
	}

	if !minimal && record[7] != "\\N" {
		runtime, err := strconv.Atoi(record[7])
		if err != nil {
			return nil, fmt.Errorf("couldn't convert string to int for runtime: %v", err)
		}
		meta.Runtime = int32(runtime)
	}

	if !minimal && record[8] != "\\N" {
		meta.Genres = strings.Split(record[8], ",")
	}

	return meta, nil
}
