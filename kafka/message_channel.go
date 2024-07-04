package main

import (
	"context"
	"github.com/IBM/sarama"
	huawei_obs "github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	"github.com/mholt/archiver/v4"
	"kafka-bucket/config"
	"kafka-bucket/obs"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultMessageChannelSize = 200000
	defaultSendFileSize       = 1048576
	defaultWorkerNumber       = 3
)

type messageChannel struct {
	msgChan chan *sarama.ConsumerMessage
}

var MsgChan *messageChannel

func init() {
	MsgChan = newMessageChannel()
	MsgChan.StartWork(defaultWorkerNumber)
}

func newMessageChannel() *messageChannel {
	return &messageChannel{
		msgChan: make(chan *sarama.ConsumerMessage, defaultMessageChannelSize),
	}
}

func (c *messageChannel) PushMessage(message *sarama.ConsumerMessage) {
	c.msgChan <- message
}

func (c *messageChannel) StartWork(workNumber int) {
	for i := 1; i <= workNumber; i++ {
		go func() {
			if err := c.consumeWorker(i); err != nil {
				log.Println(err)
				return
			}
		}()
	}
}

func (c *messageChannel) consumeWorker(workerId int) error {
	var file *os.File
	filePath := "./file" + strconv.Itoa(workerId)
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer CloseFileWithPanic(file)

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

func sendFileToObs(filePath string, workerId int) (err error) {
	compressFileName, err := compressFile(filePath, workerId)
	if err != nil {
		return err
	}
	file, err := os.Open("./" + compressFileName)
	if err != nil {
		return err
	}
	defer CloseFileWithPanic(file)
	formatPath := time.Now().Format("year=2006/month=01/day=02/hour=15")
	input := &huawei_obs.PutObjectInput{}
	input.Bucket = config.ObsConfig.Bucket
	input.Key = formatPath
	input.Body = file

	//error retry
	cnt := 3
	for cnt > 0 {
		_, err = obs.Client.PutObject(input)
		if err != nil {
			cnt--
		} else {
			break
		}
	}
	return
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
	defer CloseFileWithPanic(out)

	format := archiver.CompressedArchive{
		Compression: archiver.Bz2{},
		Archival:    archiver.Tar{},
	}
	err = format.Archive(context.Background(), out, files)
	if err != nil {
		return "", err
	}
	return filename, nil
}

func CloseFileWithPanic(file *os.File) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("close file panic: ", r)
		}
	}()
	err := file.Close()
	if err != nil {
		log.Panicln(err)
	}
}
