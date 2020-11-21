package main

import (
	"encoding/csv"
	"encoding/json"
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
)

var (
	tsvPath    = flag.String("tsvPath", "", `Path to the "data.tsv" file that's inside the "title.basics.tsv.gz" archive`)
	badgerPath = flag.String("badgerPath", "", "Path to the directory with the BadgerDB files")
	limit      = flag.Int("limit", 0, "Limit the number of rows to process (excluding the header row)")
)

var tabRune, _ = utf8.DecodeRuneInString("\t")

// Meta is the metadata of a movie or TV show
type Meta struct {
	ID            string   `json:"id"`        // IMDb ID, including "tt" prefix
	TitleType     string   `json:"titleType"` // E.g. movie, short, tvseries, tvepisode, video, etc.
	PrimaryTitle  string   `json:"primaryTitle"`
	OriginalTitle string   `json:"originalTitle"` // Only filled if different from the primary title
	IsAdult       bool     `json:"isAdult"`
	StartYear     int      `json:"startYear"` // Start year for TV shows, release year for movies. Can be 0.
	EndYear       int      `json:"endYear"`   // Only relevant for TV shows
	Runtime       int      `json:"runtime"`   // In minutes. Can be 0.
	Genres        []string `json:"genres"`    // Up to three genres. Can be empty.
}

func main() {
	flag.Parse()

	// CLI argument check
	if *tsvPath == "" {
		log.Fatalln(`Missing CLI argument "-tsvPath"`)
	}
	if *badgerPath == "" {
		log.Fatalln(`Missing CLI argument "-badgerPath"`)
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

	opts := badger.DefaultOptions(*badgerPath).
		WithLoggingLevel(badger.WARNING)
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatalf("Couldn't open BadgerDB: %v\n", err)
	}
	defer db.Close()

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

		m, err := toMeta(record)
		if err != nil {
			log.Fatalf("Couldn't create Meta from record at row %v: %#v: %v\n", i, record, err)
		}

		mBytes, err := json.Marshal(m)
		if err != nil {
			log.Fatalf("Couldn't marshal Meta to JSON at row %v: %+v: %v\n", i, m, err)
		}

		err = db.Update(func(txn *badger.Txn) error {
			return txn.Set([]byte(m.ID), mBytes)
		})
		if err != nil {
			log.Fatalf("Couldn't write marshalled Meta to database at row %v: %+v: %v\n", i, m, err)
		}

		// Including the header, we've processed i+1 at this point, but it's only going to be incremented at the beginning of the next iteration.
		if (i+1)%1000 == 0 {
			log.Printf("Processed %v rows\n", i+1)
		}
	}
	end := time.Now()
	log.Printf("Processing finished. Processed %v rows\n", i)
	log.Printf("Processing took %v\n", end.Sub(start))
}

// toMeta converts a TSV record into a Meta object.
func toMeta(record []string) (Meta, error) {
	meta := Meta{}

	meta.ID = record[0]

	meta.TitleType = record[1]

	meta.PrimaryTitle = record[2]

	if record[3] != meta.PrimaryTitle {
		meta.OriginalTitle = record[3]
	}

	if record[4] == "1" {
		meta.IsAdult = true
	}

	if record[5] != "\\N" {
		startYear, err := strconv.Atoi(record[5])
		if err != nil {
			return Meta{}, fmt.Errorf("couldn't convert string to int for startYear: %v", err)
		}
		meta.StartYear = startYear
	}

	if record[6] != "\\N" {
		endYear, err := strconv.Atoi(record[6])
		if err != nil {
			return Meta{}, fmt.Errorf("couldn't convert string to int for endYear: %v", err)
		}
		meta.EndYear = endYear
	}

	if record[7] != "\\N" {
		runtime, err := strconv.Atoi(record[7])
		if err != nil {
			return Meta{}, fmt.Errorf("couldn't convert string to int for runtime: %v", err)
		}
		meta.Runtime = runtime
	}

	if record[8] != "\\N" {
		meta.Genres = strings.Split(record[8], ",")
	}

	return meta, nil
}
