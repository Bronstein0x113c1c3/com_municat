package main

import (
	"client/networking"
	"client/step"
	"fmt"
	"log"
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




	log.Println("initiating the signal....")
	signal_chan, command_chan, ctx_signal, ctx_client, ctx1, ctx2, stop, cancel1, cancel2 := step.Signal(passcode)
	defer cancel1()
	defer cancel2()
	log.Println("signal initiation done, connecting to the server....")




	client, err := networking.InitV3("192.168.1.2:8080", "caller", ctx_client)
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
