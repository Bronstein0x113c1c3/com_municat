package main

import (
	"fmt"
	"log"
	"server/step"
	"strings"
	// "github.com/google/uuid"
)

func main() {
	fmt.Print("port: ")
	var port int
	fmt.Scanln(&port)
	fmt.Print("use http3, if yes, press something contain y or Y, no otherwise: ")
	var http3 bool
	var x string
	fmt.Scanln(&x)
	if strings.Contains(x, "Y") || strings.Contains(x, "y") {
		http3 = true
	} else {
		http3 = false
	}
	defer log.Println("turn off successfully!!!")
	log.Println("initiating the server....")
	serv := step.ServInit("", port)
	defer func() {
		close(serv.Input)
		close(serv.Incoming)
		close(serv.Leaving)
	}()

	helper, lis, err := step.ConnInit(serv, http3)
	if err != nil {
		log.Fatalf("server initation got error: %v \n", err)

	}

	log.Println("connection initiated, creating signal!!")
	ctx, ctx1, stop, cancel := step.SignalInit()
	defer stop()
	defer cancel()

	log.Println("server initiated, happy serving! wait for turning off!!!")
	step.Serve(serv, ctx1, lis, helper)
	<-ctx.Done()
	stop()

	defer helper.GracefulStop()
}
