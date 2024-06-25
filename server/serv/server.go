package serv

import (
	"fmt"
	pb "server/protobuf"
	"server/types"
	"sync"

	uuid "github.com/satori/go.uuid"

)

type Serv struct {
	host     string
	port     int
	Input    chan *types.Chunk
	Output   sync.Map
	Incoming chan uuid.UUID
	Leaving  chan uuid.UUID
	pb.UnimplementedCallingServer
}

func New(host string, port int, input chan *types.Chunk) *Serv {
	return &Serv{
		host:     host,
		port:     port,
		Input:    input,
		Incoming: make(chan uuid.UUID),
		Leaving:  make(chan uuid.UUID),
	}
}
func (s *Serv) String() string {
	return fmt.Sprintf("%v:%v", s.host, s.port)
}
