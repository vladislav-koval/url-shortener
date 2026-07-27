package gokafka

type Writer interface {
	AsyncWriteMessage(message Message) bool
}
