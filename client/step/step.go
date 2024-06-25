package step

import (
	"client/control"
	"client/input"
	"client/output"
	pb "client/protobuf"
	"client/transmitting"
	"context"
	"os"
	"os/signal"
	"sync"

	"google.golang.org/grpc/metadata"
)

func Signal(passcode string) (chan struct{}, chan string, context.Context, context.Context, context.Context, context.Context, context.CancelFunc, context.CancelFunc, context.CancelFunc) {
	signal_chan := make(chan struct{})

	command_chan := make(chan string)
	ctx_signal, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)

	md_data := metadata.Pairs("passcode", passcode)
	ctx_client := metadata.NewOutgoingContext(ctx_signal, md_data)
	ctx1, cancel1 := context.WithCancel(ctx_signal)
	// defer cancel1()
	ctx2, cancel2 := context.WithCancel(ctx_signal)
	// defer cancel2()
	return signal_chan, command_chan, ctx_signal, ctx_client, ctx1, ctx2, stop, cancel1, cancel2
}
func TestConn(client pb.Calling_VoIPClient) error {
	client.Send(&pb.ClientREQ{
		Request: &pb.ClientREQ_Signal{
			Signal: &pb.ClientSignal{},
		},
	})
	_, err := client.Recv()
	return err
}

func IOinit(channel int, sample_rate float32, buffer_size int, data_length int, wg *sync.WaitGroup) (*input.Input, *output.Output, chan []byte, error) {
	input, err := input.InputInit(channel, sample_rate, buffer_size, data_length, wg)
	if err != nil {
		return nil, nil, nil, err
	}
	data_chan := make(chan []byte, 1000)
	// defer close(data_chan)
	output, err := output.OutputInit(channel, sample_rate, buffer_size, data_chan)
	if err != nil {
		close(data_chan)
		return nil, nil, nil, err
	}
	return input, output, data_chan, nil
}

func Serve(input *input.Input, output *output.Output, data_chan chan []byte, signal_chan chan struct{}, command_chan chan string,
	client pb.Calling_VoIPClient, ctx1, ctx2 context.Context, name string, stop context.CancelFunc) {
	go input.Process()
	go transmitting.Send(input.GetChannel(), client, name)
	go transmitting.Receive(data_chan, ctx1, client)
	go output.Process()
	go control.Command(signal_chan, command_chan)
	go control.Control(ctx2, command_chan, signal_chan, input, output, stop)
}
