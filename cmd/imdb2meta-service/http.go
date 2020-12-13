package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/deflix-tv/imdb2meta/pb"
)

var healthHandler fiber.Handler = func(c *fiber.Ctx) error {
	return c.SendString("OK")
}

func createMetaHandler(metaStore *metaStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		if id == "" {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		metaBytes, err := metaStore.Get(id)
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
