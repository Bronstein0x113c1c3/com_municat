package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"server/interceptor"
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

			defer log.Printf("%v is deleted!!! \n", key.(uuid.UUID).String())
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
			<-interceptor.ConnLimit
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
	defer log.Println("turn off successfully!!!")
	log.Println("initiating the server....")
	input := make(chan *types.Chunk, 1024)
	serv := serv.New("", 8080, input)
	defer func() {
		close(serv.Input)
		close(serv.Incoming)
		close(serv.Leaving)
	}()
	helper := grpc.NewServer(grpc.ChainStreamInterceptor(interceptor.Limiting, interceptor.ChannelFinding(serv)))
	pb.RegisterCallingServer(helper, serv)
	lis, err := networking.NewConnHTTP3("caller", fmt.Sprintf("%v", serv))
	if err != nil {
		log.Printf("server initation got error: %v \n", err)
	}
	// lis, err := networking.NewConnHTTP2(fmt.Sprintf("%v", serv))
	log.Println("connection initiated, creating signal!!")
	//create signal
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()
	ctx1, cancel := context.WithCancel(ctx)
	defer cancel()
	if err != nil {
		return
	}
	log.Println("server initiated, happy serving! wait for turning off!!!")
	go control(serv, ctx1)
	go helper.Serve(lis)

	<-ctx.Done()
	stop()

	defer helper.GracefulStop()
}
