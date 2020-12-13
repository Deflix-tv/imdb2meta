package main

import (
	"context"
	"log"

	"github.com/dgraph-io/badger/v2"
	"go.etcd.io/bbolt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/deflix-tv/imdb2meta/pb"
)

// grpcServer is used to implement imdb2meta.MetaFetcher.
type grpcServer struct {
	pb.UnimplementedMetaFetcherServer
	badgerDB *badger.DB
	boltDB   *bbolt.DB
}

func createGRPCserver(badgerDB *badger.DB, boltDB *bbolt.DB) *grpcServer {
	return &grpcServer{
		badgerDB: badgerDB,
		boltDB:   boltDB,
	}
}

// Get implements imdb2meta.MetaFetcher
func (s *grpcServer) Get(ctx context.Context, in *pb.MetaRequest) (*pb.Meta, error) {
	id := in.Id

	var err error
	var metaBytes []byte

	if s.badgerDB != nil {
		err = s.badgerDB.View(func(txn *badger.Txn) error {
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
		err = s.boltDB.View(func(tx *bbolt.Tx) error {
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
			return nil, status.Error(codes.NotFound, err.Error())
		}
		log.Printf("Couldn't get data from DB: %v\n", err)
		// Note: Don't expose internal error details like DB file locations to clients
		return nil, status.Error(codes.Internal, "Couldn't get data from DB")
	}

	meta := &pb.Meta{}
	err = proto.Unmarshal(metaBytes, meta)
	if err != nil {
		log.Printf("Couldn't unmarshal protocol buffer into object: %v\n", err)
		return nil, status.Error(codes.Internal, "Couldn't unmarshal protocol buffer into object")
	}

	return meta, nil
}
