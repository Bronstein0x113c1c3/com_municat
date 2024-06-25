package step

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"server/control"
	"server/interceptor"
	"server/networking"
	pb "server/protobuf"
	"server/serv"
	"server/types"

	"google.golang.org/grpc"
)

func ServInit(host string, port int) *serv.Serv {
	input := make(chan *types.Chunk, 1024)
	return serv.New("", 8080, input)
}
func ConnInit(serv *serv.Serv, http3 bool) (*grpc.Server, net.Listener, error) {
	helper := grpc.NewServer(grpc.ChainStreamInterceptor(interceptor.Limiting, interceptor.AuthenClone, interceptor.ChannelFinding(serv)))
	pb.RegisterCallingServer(helper, serv)
	if http3 {
		lis, err := networking.NewConnHTTP3("caller", fmt.Sprintf("%v", serv))
		if err != nil {
			return nil, nil, err
		}
		return helper, lis, nil
	} else {
		lis, err := networking.NewConnHTTP2(fmt.Sprintf("%v", serv))
		if err != nil {
			return nil, nil, err
		}
		return helper, lis, nil
	}
}

func SignalInit() (context.Context, context.Context, context.CancelFunc, context.CancelFunc) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	// defer stop()
	ctx1, cancel := context.WithCancel(ctx)
	// defer cancel()
	return ctx, ctx1, stop, cancel
}

func Serve(serv *serv.Serv, ctx context.Context, lis net.Listener, helper *grpc.Server) {
	go control.Control(serv, ctx)
	go helper.Serve(lis)
}
