package nsqlogrus

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/sirupsen/logrus"
)

type NewHookFunc func(client *nsq.Producer, host, topic string, level logrus.Level) (*ElasticHook, error)

func TestSyncHook(t *testing.T) {
	hookTest(NewNsqHook, "sync-log", t)
}

func TestAsyncHook(t *testing.T) {
	hookTest(NewAsyncNsqHook, "async-log", t)
}

func hookTest(hookfunc NewHookFunc, topic string, t *testing.T) {
	config := nsq.NewConfig()
	client, err := nsq.NewProducer("127.0.0.1:4150", config)
	if err != nil {
		log.Panic(err)
	}
	hook, err := hookfunc(client, "localhost", logrus.DebugLevel, topic)
	if err != nil {
		log.Panic(err)
	}
	logrus.AddHook(hook)

	samples := 100
	for index := 0; index < samples; index++ {
		logrus.Infof("Hustej msg %d", time.Now().Unix())
	}

	// Allow time for data to be processed.
		consumer, err := nsq.NewConsumer(topic, "", config)
	if err != nil {
		log.Panic(err)
	}
	consumer.ConnectToNSQD("127.0.0.1:4150")

	var count int64
	consumer.AddHandler(func(message *nsq.Message) error){
		count+=1
	})
	time.Sleep(100 * time.Second)
	if count != int64(samples) {
		t.Errorf("Not all logs pushed to elastic: expected %d got %d", samples, searchResult.TotalHits())
		t.FailNow()
	}
}
