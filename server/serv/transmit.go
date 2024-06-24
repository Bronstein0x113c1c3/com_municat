package serv

import (
	pb "server/protobuf"
	"server/types"

	uuid "github.com/satori/go.uuid"
)

//this go file is for transmitting

func receive(conn pb.Calling_VoIPServer, data_chan chan *types.Chunk, id uuid.UUID, signal chan struct{}) {
	defer close(signal)
	for {
		data, err := conn.Recv()
		if err != nil {
			// log.Println("receiving side...")
			return
		}

		//for testing...
		/*
			if the request is signal -> return the signal as successully for the test
			else, just process the sound

			cuz, these could do <=> the connection is initiated!!!!!
		*/

		c := &types.Chunk{
			ID:    id,
			Name:  data.GetName(),
			Chunk: data.GetChunk(),
		}

		data_chan <- c
	}
}
func send(conn pb.Calling_VoIPServer, data_chan chan *types.Chunk, id uuid.UUID, signal chan struct{}) {
	defer close(signal)
	for {
		data, ok := <-data_chan
		if !ok {
			return
		}
		conn.Send(
			&pb.ServerRES{
				Msg: &pb.ClientMSG{
					Chunk: data.Chunk,
					Name:  data.Name,
				},
				Id: id.String(),
			},
		)
	}
}
