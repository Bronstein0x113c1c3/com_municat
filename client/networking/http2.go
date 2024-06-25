package networking

import (
	pb "client/protobuf"
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitV2(host string, ctx context.Context) (pb.Calling_VoIPClient, error) {
	grpcOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	conn, err := grpc.NewClient(host, grpcOpts...)
	if err != nil {
		return nil, err
	}
	client, err := pb.NewCallingClient(conn).VoIP(ctx)
	if err != nil {
		return nil, err
	}
	return client, nil
}
