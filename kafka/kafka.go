package main

// SIGUSR1 toggle the pause/resume consumption
import (
	"context"
	"errors"
	"kafka-bucket/config"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/IBM/sarama"
)

// Sarama configuration options
var (
	brokers  = config.KafkaConfig.BootstrapServers
	group    = config.KafkaConfig.Consumer.GroupId
	topics   = config.KafkaConfig.Topic
	assignor = "roundrobin"
)

func main() {
	keepRunning := true
	log.Println("Starting a new Sarama consumer")
	kafkaConfig := sarama.NewConfig()
	if config.KafkaConfig.Consumer.AutoOffsetReset == "latest" {
		kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest
	}
	kafkaConfig.Consumer.Offsets.AutoCommit.Enable = false

	switch assignor {
	case "sticky":
		kafkaConfig.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategySticky()}
	case "roundrobin":
		kafkaConfig.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	case "range":
		kafkaConfig.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRange()}
	default:
		log.Panicf("Unrecognized consumer group partition assignor: %s", assignor)
	}

	consumer := Consumer{
		ready: make(chan bool),
	}

	ctx, cancel := context.WithCancel(context.Background())
	client, err := sarama.NewConsumerGroup(brokers, group, kafkaConfig)
	if err != nil {
		log.Panicf("Error creating consumer group client: %v", err)
	}

	consumptionIsPaused := false
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if err := client.Consume(ctx, strings.Split(topics, ","), &consumer); err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					return
				}
				log.Panicf("Error from consumer: %v", err)
			}

			if ctx.Err() != nil {
				return
			}
			consumer.ready = make(chan bool)
		}
	}()

	<-consumer.ready
	log.Println("Sarama consumer up and running!...")

	sigusr1 := make(chan os.Signal, 1)
	// in windows , there is no signal sigusr1
	//signal.Notify(sigusr1, syscall.SIGUSR1)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	for keepRunning {
		select {
		case <-ctx.Done():
			log.Println("terminating: context cancelled")
			keepRunning = false
		case <-sigterm:
			log.Println("terminating: via signal")
			keepRunning = false
		case <-sigusr1:
			toggleConsumptionFlow(client, &consumptionIsPaused)
		}
	}
	cancel()
	wg.Wait()
	if err = client.Close(); err != nil {
		log.Panicf("Error closing client: %v", err)
	}
}

func toggleConsumptionFlow(client sarama.ConsumerGroup, isPaused *bool) {
	if *isPaused {
		client.ResumeAll()
		log.Println("Resuming consumption")
	} else {
		client.PauseAll()
		log.Println("Pausing consumption")
	}
	*isPaused = !*isPaused
}
