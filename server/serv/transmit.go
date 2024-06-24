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
