package config

import (
	"github.com/spf13/viper"
	"os"
)

type kafkaConfig struct {
	BootstrapServers []string
	Topic            string
	Consumer         consumer
}

type consumer struct {
	GroupId         string
	MaxPollRecords  int
	AutoOffsetReset string
}

type obsConfig struct {
	Endpoint     string
	Bucket       string
	SendFileSize int
	FilePrefix   string
}

var KafkaConfig kafkaConfig
var ObsConfig obsConfig

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("D:\\goland\\kafka-bucket\\config")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	initKafka()
	initObs()
}

func initKafka() {
	KafkaConfig.BootstrapServers = viper.GetStringSlice("kafka.bootstrap_servers")
	KafkaConfig.Topic = viper.GetString("kafka.topic")
	KafkaConfig.Consumer.MaxPollRecords = viper.GetInt("kafka.consumer.max_poll_records")
	KafkaConfig.Consumer.AutoOffsetReset = viper.GetString("kafka.consumer.auto_offset_reset")
}

func initObs() {
	ObsConfig.FilePrefix = os.Getenv("MY_POD_NAME")
	ObsConfig.Endpoint = viper.GetString("obs.endpoint")
	ObsConfig.Bucket = viper.GetString("obs.bucket")
	ObsConfig.SendFileSize = viper.GetInt("obs.send_file_size")
}
