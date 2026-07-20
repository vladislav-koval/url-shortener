package kafka

import "context"

type Writer interface {
	WriteMessage(ctx context.Context, key, value []byte) error
	Close() error
}
