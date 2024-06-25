package serv

import (
	"log"
	pb "server/protobuf"
	"server/types"

	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Serv) VoIP(conn pb.Calling_VoIPServer) error {
	// id, err := s.In()
	// if err != nil {
	// 	return status.Error(codes.Internal, "create got problem...")
	// }
	id := conn.Context().Value(types.T("channel")).(uuid.UUID)
	res, ok := s.Output.Load(id)
	if !ok {
		return status.Error(codes.Internal, "losing channel...")
	}
	channel := *res.(*types.Conn)

	sig1 := make(chan struct{})
	sig2 := make(chan struct{})
	//receive the byte to send to the channel
	go receive(conn, s.Input, id, sig1)
	//get the data then send to the client
	go send(conn, channel, id, sig2)
	for {
		select {
		case <-sig1:
			log.Println("Disconnected....")
			s.out(id, false)
			return nil
		case <-sig2:
			log.Println("Forcing closed!!!!")
			s.out(id, true)
			return nil
		}
	}
}
