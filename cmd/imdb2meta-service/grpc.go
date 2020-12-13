package main

import (
	"context"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/deflix-tv/imdb2meta/pb"
)

// grpcServer is used to implement imdb2meta.MetaFetcher.
type grpcServer struct {
	pb.UnimplementedMetaFetcherServer
	metaStore *metaStore
}

func createGRPCserver(metaStore *metaStore) *grpcServer {
	return &grpcServer{
		metaStore: metaStore,
	}
}

// Get implements imdb2meta.MetaFetcher.
func (s *grpcServer) Get(ctx context.Context, in *pb.MetaRequest) (*pb.Meta, error) {
	id := in.Id
	metaBytes, err := s.metaStore.Get(id)
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
