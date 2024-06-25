package networking

import (
	pb "client/protobuf"
	"context"
	"crypto/tls"

	grpcquic "github.com/coremedic/grpc-quic"
	"google.golang.org/grpc"
)

func InitV3(host string, passcode string, ctx context.Context) (pb.Calling_VoIPClient, error) {
	passcodes := []string{}
	passcodes = append(passcodes, passcode)
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         passcodes,
	}
	creds := grpcquic.NewCredentials(tlsConfig)

	// Connect to gRPC Service Server
	dialer := grpcquic.NewQuicDialer(tlsConfig)
	// grpc
	grpcOpts := []grpc.DialOption{
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(creds),
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
