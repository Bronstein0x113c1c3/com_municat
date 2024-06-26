package transmitting

import (
	pb "client/protobuf"
	"context"
)

func Send(data_chan chan []byte, client pb.Calling_VoIPClient, name string) {
	// data_chan := input.GetChannel()
	for data := range data_chan {
		// client.Send(&pb.ClientMSG{
		// 	Chunk: data,
		// 	Name:  name,
		// })
		client.Send(&pb.ClientREQ{
			Request: &pb.ClientREQ_Message{
				Message: &pb.ClientMSG{
					Chunk: data,
					Name:  name,
				},
			},
		})
	}
}
func Receive(data_chan chan []byte, ctx context.Context, client pb.Calling_VoIPClient, stop context.CancelFunc) {

	for {
		select {
		case <-ctx.Done():
			return
		default:
			segment, err := client.Recv()
			if segment.GetSignal() != nil {
				continue
			}
			data := segment.GetMessage()
			if err != nil {
				stop()
				return
			}
			data_chan <- data.Msg.Chunk
		}
	}
}
