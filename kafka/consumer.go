package main

import (
	"github.com/IBM/sarama"
	"log"
	"os"
	"time"
)

// Consumer represents a Sarama consumer group consumer
type Consumer struct {
	ready chan bool
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(consumer.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
// Once the Messages() channel is closed, the Handler must finish its processing
// loop and exit.
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				log.Printf("message channel was closed")
				return nil
			}
			//message.Value, message.Timestamp, message.Topic
			session.MarkMessage(message, "has been consumed. time: "+time.Now().Format(time.DateTime))
			session.Commit()
		case <-session.Context().Done():
			return nil
		}
	}
}

func writeMessage(message *sarama.ConsumerMessage) error {
	file, err := os.OpenFile("./compress_message", os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	_, err = file.Write(message.Value)
	if err != nil {
		return err
	}
	if err = file.Sync(); err != nil {
		return err
	}
	return nil
}

func compress() {

}

func sendToObs() {

}
