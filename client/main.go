package main

import (
	"client/input"
	"client/networking"
	"client/output"
	pb "client/protobuf"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"

	"google.golang.org/grpc/metadata"
)

func main() {

	fmt.Print("if you want to continue, press your name, otherwise, let it blank: ")
	var name string
	fmt.Scanln(&name)
	if name == "" {
		return
	}
	fmt.Print("please press the passcode: ")
	var passcode string
	fmt.Scanln(&passcode)
	log.Println("io done, initiating the signal....")
	ctx_signal, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	md_data := metadata.Pairs("passcode", passcode)
	// ctx := context.WithValue(ctx_signal, "passcode", passcode)
	ctx_client := metadata.NewOutgoingContext(ctx_signal, md_data)

	ctx1, cancel1 := context.WithCancel(ctx_signal)
	defer cancel1()
	log.Println("connecting to the server....")
	client, err := networking.InitV3("192.168.1.2:8080", "caller", ctx_client)
	// client, err := init_the_client_v2("192.168.1.2:8080", ctx)
	if err != nil {
		log.Fatalf("error before connecting: %v \n", err)
	}
	//connection testing....

	client.Send(&pb.ClientREQ{
		Request: &pb.ClientREQ_Signal{
			Signal: &pb.ClientSignal{},
		},
	})
	_, err = client.Recv()
	if err != nil {
		log.Fatalf("error after connection: %v \n", err)
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

	fmt.Println("please press ctrl+c anytime you want to stop, please :)")
	go input.Process()
	go func() {
		data_chan := input.GetChannel()
		for data := range data_chan {
			// client.Send(&pb.ClientMSG{
			// 	Chunk: data,
			// 	Name:  name,
			// })
			client.Send(&pb.ClientREQ{
				Request: &pb.ClientREQ_Message{
					Message: &pb.ClientMSG{
						Chunk: data,
						Name:  name,
					},
				},
			})
		}
		// for data:= range
	}()
	go func() {

		for {
			select {
			case <-ctx1.Done():
				return
			default:
				segment, err := client.Recv()
				if segment.GetSignal() != nil {
					continue
				}
				data := segment.GetMessage()
				if err != nil {
					stop()
					return
				}
				data_chan <- data.Msg.Chunk
			}
		}
	}()
	go output.Process()
	<-ctx_signal.Done()
	stop()
	go input.Close()
	go output.Close()
	defer wg.Wait()
}
