package types

import uuid "github.com/satori/go.uuid"

type Chunk struct {
	ID    uuid.UUID
	Name  string
	Chunk []byte
}

type Conn chan *Chunk
