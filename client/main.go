package main

import (
	"client/input"
	"client/output"
	pb "client/protobuf"
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"

	grpcquic "github.com/coremedic/grpc-quic"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func init_the_client_v3(host string, passcode string, ctx context.Context) (pb.Calling_VoIPClient, error) {
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
func init_the_client_v2(host string, ctx context.Context) (pb.Calling_VoIPClient, error) {
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

func main() {
	log.Println("io done, initiating the signal....")
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()
	ctx1, cancel1 := context.WithCancel(ctx)
	defer cancel1()
	log.Println("connecting to the server....")
	client, err := init_the_client_v3("192.168.1.2:8080", "caller", ctx)
	// client, err := init_the_client_v2("192.168.1.2:8080", ctx)

	if err != nil {
		log.Fatalln(err)
	}
	defer client.CloseSend()
	wg := &sync.WaitGroup{}

	log.Println("connected, init the i/o...")
	input, err := input.InputInit(DefaultChannels, DefaultSampleRate, DefaultFrameSize, DefaultOpusDataLength, wg)
	if err != nil {
		return
	}
	data_chan := make(chan []byte, 1000)
	// defer close(data_chan)
	output, err := output.OutputInit(DefaultChannels, DefaultSampleRate, DefaultFrameSize, data_chan)
	if err != nil {
		return
	}

	wg.Add(1)
	log.Println("signal done, start processing....")

	fmt.Print("your name?: ")
	var name string
	fmt.Scanln(&name)
	fmt.Println("please press ctrl+c anytime you want to stop, please :)")
	go input.Process()
	go func() {
		data_chan := input.GetChannel()
		for data := range data_chan {
			client.Send(&pb.ClientMSG{
				Chunk: data,
				Name:  name,
			})
		}
		// for data:= range
	}()
	go func() {

		for {
			select {
			case _, ok := <-ctx1.Done():
				if !ok {
					return
				}
			default:
				data, err := client.Recv()
				if err != nil {
					stop()
					return
				}
				data_chan <- data.Msg.Chunk
			}
		}
	}()
	go output.Process()
	<-ctx.Done()
	stop()
	go input.Close()
	go output.Close()
	defer wg.Wait()
}
