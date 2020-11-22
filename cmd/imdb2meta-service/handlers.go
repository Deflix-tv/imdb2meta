package main

import (
	"errors"
	"log"

	"github.com/deflix-tv/imdb2meta/pb"
	"github.com/dgraph-io/badger/v2"
	"github.com/gofiber/fiber/v2"
	"go.etcd.io/bbolt"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var (
	errNotFound = errors.New("Not found")
)

var healthHandler fiber.Handler = func(c *fiber.Ctx) error {
	return c.SendString("OK")
}

func createMetaHandler(badgerDB *badger.DB, boltDB *bbolt.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		if id == "" {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		var err error
		var metaBytes []byte

		if badgerDB != nil {
			err = badgerDB.View(func(txn *badger.Txn) error {
				item, err := txn.Get([]byte(id))
				if err != nil {
					if err == badger.ErrKeyNotFound {
						return errNotFound
					}
					return err
				}
				metaBytes, err = item.ValueCopy(nil)
				return err
			})
		} else {
			err = boltDB.View(func(tx *bbolt.Tx) error {
				txBytes := tx.Bucket(imdbBytes).Get([]byte(id))
				if txBytes == nil {
					return errNotFound
				}
				copy(metaBytes, txBytes)
				return nil
			})
		}
		if err != nil {
			if err == errNotFound {
				log.Printf("Key not found in DB: %v\n", err)
				return c.SendStatus(fiber.StatusNotFound)
			}
			log.Printf("Couldn't get data from DB: %v\n", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		meta := &pb.Meta{}
		err = proto.Unmarshal(metaBytes, meta)
		if err != nil {
			log.Printf("Couldn't unmarshal protocol buffer into object: %v\n", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		metaJSON, err := protojson.Marshal(meta)
		if err != nil {
			log.Printf("Couldn't marshal object into JSON: %v\n", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
		return c.Send(metaJSON)
	}
}
