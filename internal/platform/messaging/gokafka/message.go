package gokafka

type Message struct {
	Key   []byte
	Value []byte

	Topic     string
	Partition int
	Offset    int64
}
