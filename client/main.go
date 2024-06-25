package main

import (
	"client/control"
	"client/input"
	"client/networking"
	"client/output"
	pb "client/protobuf"
	"client/transmitting"
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

	log.Println("initiating the signal....")
	signal_chan := make(chan struct{})
	command_chan := make(chan string)
	ctx_signal, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()
	md_data := metadata.Pairs("passcode", passcode)
	ctx_client := metadata.NewOutgoingContext(ctx_signal, md_data)
	ctx1, cancel1 := context.WithCancel(ctx_signal)
	defer cancel1()
	ctx2, cancel2 := context.WithCancel(ctx_signal)
	defer cancel2()
	log.Println("signal initiation done, connecting to the server....")

	client, err := networking.InitV3("192.168.1.2:8080", "caller", ctx_client)
	if err != nil {
		log.Fatalf("error before connecting: %v \n", err)
	}
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
	log.Println("io done, start processing....")
	fmt.Println("please press ctrl+c anytime you want to stop, please :)")

	go input.Process()
	go transmitting.Send(input.GetChannel(), client, name)
	go transmitting.Receive(data_chan, ctx1, client)
	go output.Process()
	go control.Command(signal_chan, command_chan)
	go control.Control(ctx2, command_chan, signal_chan, input, output, stop)

	<-ctx_signal.Done()
	defer wg.Wait()
}
