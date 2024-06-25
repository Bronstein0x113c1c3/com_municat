package interceptor

import (
	"log"
	"server/types"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func AuthenClone(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	// p, ok := peer.FromContext(ss.Context())
	// log.Printf("%v is called... \n", info.FullMethod)
	// if !ok {
	// log.Println("Cannot get any context.....")
	// }
	// log.Printf("%v wants to connect....: ", p.Addr.String())

	// a := 10 - len(ConnLimit)
	// if a <= 0 {
	// log.Printf("remaining %v connections... \n", a)
	// log.Println("Out of slots")
	// return status.Error(codes.ResourceExhausted, "out of slots, cancelled!!!")
	// }
	// log.Println("Accepted!!!")
	// ConnLimit <- struct{}{}
	// a = 10 - len(ConnLimit)
	// log.Printf("remaining %v connections... \n", a)
	md, ok := metadata.FromIncomingContext(ss.Context())
	if !ok {
		log.Println("does not have any context....")
	}
	passcode := md.Get("passcode")[0]
	if passcode != types.Passcode {
		log.Println("wrong passcode!!!!")
		<-ConnLimit
		return status.Error(codes.Unauthenticated, "out of slots, cancelled!!!")
	}
	// log.Println(passcode)
	handler(srv, ss)
	return nil
}
