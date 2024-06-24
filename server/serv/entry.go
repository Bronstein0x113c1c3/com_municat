package serv

import (
	"errors"
	"log"
	"server/types"

	xuuid "github.com/satori/go.uuid"
)

//this go file is for controlling the Output channel for in/out

func (s *Serv) in() (xuuid.UUID, error) {

	id := xuuid.NewV4()
	data := make(types.Conn, 1000)
	_, stored := s.Output.LoadOrStore(id, &data)
	if stored {
		return xuuid.Nil, errors.New("error in creating channel")
	}
	s.Incoming <- id
	return id, nil
}
func (s *Serv) out(id xuuid.UUID, closed bool) {
	log.Printf("Delete for %v is called \n", id.String())
	defer log.Printf("%v is released \n", id.String())
	if !closed {
		s.Leaving <- id
		log.Println("Deletion requested!!!")
		return
	}
	
	log.Printf("Released %v!!!", id.String())
}
