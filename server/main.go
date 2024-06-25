package main

import (
	"log"
	"server/step"
	// "github.com/google/uuid"
)

func main() {
	defer log.Println("turn off successfully!!!")
	log.Println("initiating the server....")
	serv := step.ServInit("", 8080)
	defer func() {
		close(serv.Input)
		close(serv.Incoming)
		close(serv.Leaving)
	}()

	helper, lis, err := step.ConnInit(serv, false)
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
