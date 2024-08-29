package commonFunctions

import (
	"context"

	"github.com/geico-private/pv-bil-frameworks/kafkamessaging"
	"github.com/segmentio/kafka-go"
)

func NewTestKafkaReader(brokers []string, groupID string, topic string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		GroupID: groupID,
		Topic:   topic,
		Dialer:  &kafka.Dialer{},
	})
}

// for integration test purpose only.
func ProcessOneMesssageFromReader(r *kafka.Reader, handler kafkamessaging.MessageHandlerFunc) error {
	m, err := r.ReadMessage(context.Background())
	if err != nil {
		return err
	}
	return handler.HandleMessage(context.Background(), mapToStreamingMessage(&m))
}

func mapToStreamingMessage(message *kafka.Message) *kafkamessaging.Message {
	streamMessage := kafkamessaging.Message{}

	for _, header := range message.Headers {
		if header.Key == "MessageId" {
			streamMessage.MessageID = string(header.Value)
			break
		}
	}
	streamMessage.Key = string(message.Key)
	streamMessage.Partition = message.Partition
	streamMessage.Offset = message.Offset
	body := string(message.Value)
	streamMessage.Body = &body

	return &streamMessage
}
