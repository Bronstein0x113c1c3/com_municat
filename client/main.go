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
	// ctx := context.WithValue(ctx_signal, "passcode", passcode)
	ctx_client := metadata.NewOutgoingContext(ctx_signal, md_data)

	ctx1, cancel1 := context.WithCancel(ctx_signal)
	defer cancel1()
	ctx2, cancel2 := context.WithCancel(ctx_signal)
	defer cancel2()
	log.Println("signal initiation done, connecting to the server....")
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
	log.Println("io done, start processing....")
	fmt.Println("please press ctrl+c anytime you want to stop, please :)")

	go input.Process()
	go transmitting.Send(input.GetChannel(), client, name)
	go transmitting.Receive(data_chan, ctx1, client)
	go output.Process()

	// go func(signal chan struct{}) {
	// 	defer close(command_chan)
	// 	for {
	// 		select {
	// 		case <-signal:
	// 			return
	// 		default:
	// 			for {
	// 				var command string
	// 				fmt.Print("what command: ")
	// 				fmt.Scanln(&command)
	// 				command_chan <- command
	// 			}
	// 		}
	// 	}
	// }(signal_chan)
	go control.Command(signal_chan, command_chan)
	// go func(ctx context.Context, command chan string, signal chan struct{}) {
	// 	defer close(signal_chan)
	// 	for {
	// 		select {
	// 		case <-ctx.Done():
	// 			log.Println("ctrl + c detected, exitting...")
	// 			go input.Close()
	// 			go output.Close()
	// 			return
	// 		case x, ok := <-command:
	// 			if !ok {
	// 				return
	// 			}
	// 			switch x {
	// 			case "mute":
	// 				log.Println("muting called")
	// 				go input.Mute()
	// 				continue
	// 			case "unmute":
	// 				log.Println("unmuting called")

	// 				go input.Play()
	// 				continue
	// 			case "stop":
	// 				log.Println("stopping called")

	// 				go input.Close()
	// 				go output.Close()
	// 				stop()
	// 				return
	// 			default:

	// 			}
	// 			// var command
	// 		}
	// 	}
	// }(ctx2, command_chan, signal_chan)
	go control.Control(ctx2, command_chan, signal_chan, input, output, stop)
	<-ctx_signal.Done()
	// stop()
	// go input.Close()
	// go output.Close()
	defer wg.Wait()
}
