package main

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/mholt/archiver/v4"
	"log"
	"os"
	"testing"
	"time"
)

func TestTimeBucket(t *testing.T) {
	print(time.Now().Format("year=2006/month=01/day=02/hour=15"))
}

func TestArchive(t *testing.T) {
	log.Println("Start")
	go func() {
		for {
			MsgChan.PushMessage(&sarama.ConsumerMessage{
				Key:   []byte("test_key \n"),
				Value: []byte("test_value \n"),
			})
		}
	}()
	select {}
}

func TestCreateArchive(t *testing.T) {
	files, err := archiver.FilesFromDisk(nil, map[string]string{
		"./file1": "file1",
	})
	filename := "compress_file.tar.bz2"
	out, err := os.Create(filename)
	if err != nil {
		t.Error(err)
	}
	defer out.Close()
	format := archiver.CompressedArchive{
		Compression: archiver.Bz2{},
		Archival:    archiver.Tar{},
	}
	err = format.Archive(context.Background(), out, files)
	if err != nil {
		t.Error(err)
	}
}

func TestFile(t *testing.T) {
	log.Println("test")
}
