package main

import (
	"client/step"
	"fmt"
	"log"
	"strings"
	"sync"
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

	fmt.Print("do you want to use http3? if yes, press something contain y or Y, no otherwise: ")
	var http3 bool
	var x string
	fmt.Scanln(&x)
	if strings.Contains(x, "Y") || strings.Contains(x, "y") {
		http3 = true
	} else {
		http3 = false
	}

	fmt.Print("where do you want to connect: ")
	var host string
	fmt.Scanln(&host)

	fmt.Println()
	log.Println("initiating the signal....")
	signal_chan, command_chan, ctx_signal, ctx_client, ctx1, ctx2, stop, cancel1, cancel2 := step.Signal(passcode)
	defer cancel1()
	defer cancel2()
	log.Println("signal initiation done, connecting to the server....")

	client, err := step.InitClient(host, ctx_client, "caller", http3)

	if err != nil {
		log.Fatalf("error before connecting: %v \n", err)
	}
	if err := step.TestConn(client); err != nil {
		log.Fatalf("error after connection: %v \n", err)
	}
	defer client.CloseSend()

	log.Println("connected, init the i/o...")
	wg := &sync.WaitGroup{}
	input, output, data_chan, err := step.IOinit(DefaultChannels, DefaultSampleRate, DefaultFrameSize, DefaultOpusDataLength, wg)
	if err != nil {
		log.Fatalf("Error encountered when initializing I/O: %v \n", err)
		// return
	}
	defer close(data_chan)
	log.Println("io done, start processing....")

	fmt.Println("please press ctrl+c anytime you want to stop, please :)")
	wg.Add(1)
	step.Serve(input, output, data_chan, signal_chan, command_chan, client, ctx1, ctx2, name, stop)
	<-ctx_signal.Done()
	defer wg.Wait()
}
