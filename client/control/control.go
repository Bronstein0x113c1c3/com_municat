package control

import (
	"client/input"
	"client/output"
	"context"
	"fmt"
	"log"
)

func Command(signal chan struct{}, command_chan chan string) {
	defer close(command_chan)
	for {
		select {
		case <-signal:
			return
		default:
			for {
				var command string
				fmt.Print("what command: ")
				fmt.Scanln(&command)
				command_chan <- command
			}
		}
	}
}
func Control(ctx context.Context, command chan string, signal chan struct{}, input *input.Input, output *output.Output, stop context.CancelFunc) {
	defer close(signal)
	for {
		select {
		case <-ctx.Done():
			log.Println("ctrl + c detected, exitting...")
			go input.Close()
			go output.Close()
			return
		case x, ok := <-command:
			if !ok {
				return
			}
			switch x {
			case "mute":
				log.Println("muting called")
				go input.Mute()
				continue
			case "unmute":
				log.Println("unmuting called")

				go input.Play()
				continue
			case "stop":
				log.Println("stopping called")
				go input.Close()
				go output.Close()
				stop()
				return
			default:

			}
			// var command
		}
	}
}
