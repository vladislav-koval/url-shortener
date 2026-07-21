package gokafka

import (
	"context"
)

type Message struct {
	Key   []byte
	Value []byte

	Topic     string
	Partition int
	Offset    int64
}

type Reader interface {
	FetchMessage(ctx context.Context) (Message, error)
	CommitMessages(ctx context.Context, msgs ...Message) error
	Close() error
}
