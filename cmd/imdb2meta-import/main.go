package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

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
	minimal      = flag.Bool("minimal", false, "Only store minimal metadata (ID, type, title, release/start year)")
)

var (
	tabRune, _ = utf8.DecodeRuneInString("\t")
	imdbBytes  = []byte("imdb") // Bucket name for bbolt
)

func main() {
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

	r := csv.NewReader(f)
	r.Comma = tabRune
	// ReuseRecord for better performance.
	// Note: In this case the slices of the returned records when reading are backed by a reused array!
	r.ReuseRecord = true
	// Required for example for row 32542
	r.LazyQuotes = true

	i := 1
	// The first row is just the headers
	_, err = r.Read()
	if err == io.EOF {
		log.Fatalf("The TSV file doesn't seem to contain any data: %v\n", err)
	}
	if err != nil {
		log.Fatalf("Couldn't read TSV row %v: %v\n", i, err)
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

	storedCount := 0
	start := time.Now()
	for ; *limit == 0 || i <= *limit; i++ {
		record, err := r.Read()
		if err == io.EOF {
			// No need to decrement i here
			break
		}
		if err != nil {
			log.Fatalf("Couldn't read TSV row %v: %v\n", i, err)
		}

		m, err := toMeta(record, *minimal)
		if err != nil {
			log.Fatalf("Couldn't create Meta from record at row %v: %#v: %v\n", i, record, err)
		}

		// Skip TV episodes if configured
		if *skipEpisodes && m.GetTitleType() == pb.TitleType_TV_EPISODE {
			continue
		}

		mBytes, err := proto.Marshal(m)
		if err != nil {
			log.Fatalf("Couldn't marshal Meta to protocol buffer at row %v: %+v: %v\n", i, m, err)
		}

		if *badgerPath != "" {
			err = badgerDB.Update(func(txn *badger.Txn) error {
				return txn.Set([]byte(m.GetId()), mBytes)
			})
		} else {
			err = boltDB.Update(func(tx *bbolt.Tx) error {
				return tx.Bucket(imdbBytes).Put([]byte(m.GetId()), mBytes)
			})
		}
		if err != nil {
			log.Fatalf("Couldn't write marshalled Meta to database at row %v: %+v: %v\n", i, m, err)
		}
		storedCount++

		// Including the header, we've processed i+1 at this point, but it's only going to be incremented at the beginning of the next iteration.
		if (i+1)%1000 == 0 {
			log.Printf("Processed %v rows, stored %v objects\n", i+1, storedCount)
		}
	}
	end := time.Now()
	log.Printf("Processing finished. Processed %v rows, stored %v objects.\n", i, storedCount)
	log.Printf("Processing took %v\n", end.Sub(start))
}

// toMeta converts a TSV record into a Meta object.
func toMeta(record []string, minimal bool) (*pb.Meta, error) {
	meta := &pb.Meta{}

	meta.Id = record[0]

	// As of 2020-11-21 this can be "movie", "short", "tvEpisode", "tvMiniSeries", "tvMovie", "tvSeries", "tvShort", "tvSpecial", "video" or "videoGame"
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
