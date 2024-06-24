package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"server/networking"
	pb "server/protobuf"
	"server/serv"
	"server/types"

	// "github.com/google/uuid"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
)

func control(serv *serv.Serv, ctx context.Context) {
	defer func() {
		log.Println("starting to close...")
		serv.Output.Range(func(key, value any) bool {
			channel := *value.(*types.Conn)
			close(channel)
			defer serv.Output.Delete(key)
			// log.Printf("We got data from: %v, %v", id, <-channel)
			return true
		})
	}()
	for {
		select {
		case <-ctx.Done():
			log.Println("cancellation is received, start to close....")
			return
		case x := <-serv.Incoming:
			log.Printf("%v is coming!!! \n", x.String())
			continue
		case x := <-serv.Leaving:
			log.Printf("%v is leaving!!! \n", x.String())
			res, ok := serv.Output.Load(x)
			if !ok {
				return
			}
			channel := *res.(*types.Conn)
			close(channel)
			serv.Output.Delete(x)
			continue
		case data, ok := <-serv.Input:
			if !ok {
				log.Println("Channel is forcibly closed")
				return
			}
			serv.Output.Range(func(key, value any) bool {
				channel := *value.(*types.Conn)
				id := key.(uuid.UUID)
				if !uuid.Equal(data.ID, id) {
					channel <- data
				}
				// log.Printf("We got data from: %v, %v", id, <-channel)
				return true
			})
		}
	}
}
func main() {
	input := make(chan *types.Chunk, 1024)
	serv := serv.New("", 8080, input)
	defer close(serv.Input)
	helper := grpc.NewServer()
	pb.RegisterCallingServer(helper, serv)
	lis, err := networking.NewConnHTTP3("caller", fmt.Sprintf("%v", serv))
	// lis, err := networking.NewConnHTTP2(fmt.Sprintf("%v", serv))

	//create signal
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()
	ctx1, cancel := context.WithCancel(ctx)
	defer cancel()
	if err != nil {
		return
	}

	go control(serv, ctx1)
	go helper.Serve(lis)

	<-ctx.Done()
	stop()

	defer helper.GracefulStop()
}
