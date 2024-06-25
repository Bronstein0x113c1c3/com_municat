package control

import (
	"context"
	"log"
	"server/interceptor"
	"server/serv"
	"server/types"

	uuid "github.com/satori/go.uuid"
)

func Control(serv *serv.Serv, ctx context.Context) {
	defer func() {
		log.Println("starting to close...")
		serv.Output.Range(func(key, value any) bool {
			channel := *value.(*types.Conn)
			close(channel)

			defer log.Printf("%v is deleted!!! \n", key.(uuid.UUID).String())
			defer serv.Output.Delete(key)
			// log.Printf("We got data from: %v, %v", id, <-channel)
			return true
		})
	}()
	for {
		select {
		case <-ctx.Done():
			log.Println("cancellation is received, start to close....")
			return
		case x := <-serv.Incoming:
			log.Printf("%v is coming!!! \n", x.String())
			continue
		case x := <-serv.Leaving:
			log.Printf("%v is leaving!!! \n", x.String())
			res, ok := serv.Output.Load(x)
			if !ok {
				return
			}
			channel := *res.(*types.Conn)
			close(channel)
			serv.Output.Delete(x)
			<-interceptor.ConnLimit
			continue
		case data, ok := <-serv.Input:
			if !ok {
				log.Println("Channel is forcibly closed")
				return
			}
			serv.Output.Range(func(key, value any) bool {
				channel := *value.(*types.Conn)
				id := key.(uuid.UUID)
				if !uuid.Equal(data.ID, id) {
					channel <- data
				}
				// log.Printf("We got data from: %v, %v", id, <-channel)
				return true
			})
		}
	}
}
