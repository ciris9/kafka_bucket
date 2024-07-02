package main

import (
	"context"
	"github.com/IBM/sarama"
	huawei_obs "github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	"github.com/mholt/archiver/v4"
	"kafka-bucket/config"
	"kafka-bucket/obs"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultMessageChannelSize = 200000
	defaultSendFileSize       = 1048576
)

type MessageChannel struct {
	msgChan chan *sarama.Message
	quit    chan bool
}

func NewMessageChannel() *MessageChannel {
	return &MessageChannel{
		msgChan: make(chan *sarama.Message, defaultMessageChannelSize),
		quit:    make(chan bool),
	}
}

func (c *MessageChannel) ConsumeWorker(workerId int) error {
	var file *os.File
	filePath := "./file" + strconv.Itoa(workerId)
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	defer file.Close()
	if err != nil {
		return err
	}

	for msg := range c.msgChan {
		fileInfo, err := file.Stat()
		if err != nil {
			return err
		}
		if fileInfo.Size() > defaultSendFileSize {
			if err := sendFileToObs(filePath, workerId); err != nil {
				return err
			}
			if err := file.Truncate(0); err != nil {
				return err
			}
		}

		if _, err = file.Write(msg.Value); err != nil {
			return err
		}
		if err := file.Sync(); err != nil {
			return err
		}
	}
	return nil
}

func sendFileToObs(filePath string, workerId int) error {
	compressFileName, err := compressFile(filePath, workerId)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(compressFileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	formatPath := time.Now().Format("year=2006/month=01/day=02/hour=15")
	input := &huawei_obs.PutObjectInput{}
	input.Bucket = config.ObsConfig.Bucket
	input.Key = formatPath
	input.Body = file
	_, err = obs.Client.PutObject(input)
	if err != nil {
		return err
	}
	return nil
}

func compressFile(filePath string, workerId int) (string, error) {
	paths := strings.Split(filePath, "/")
	files, err := archiver.FilesFromDisk(nil, map[string]string{
		filePath: paths[len(paths)-1],
	})
	filename := "compress_file" + strconv.Itoa(workerId) + ".tar.bz2"
	out, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	format := archiver.CompressedArchive{
		Compression: archiver.Bz2{},
		Archival:    archiver.Zip{},
	}
	err = format.Archive(context.Background(), out, files)
	return filename, nil
}
