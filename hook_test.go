package nsqlogrus

import (
	"log"
	"testing"
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/sirupsen/logrus"
)

type NewHookFunc func(client *nsq.Producer, host, topic string, level logrus.Level) (*NsqHook, error)

func TestSyncHook(t *testing.T) {
	hookTest(NewNsqHook, "sync-log", t)
}

func TestAsyncHook(t *testing.T) {
	hookTest(NewAsyncNsqHook, "async-log", t)
}

type nsqHandle struct {
	ch chan int
}

func (h *nsqHandle) HandleMessage(m *nsq.Message) error {
	h.ch <- 0
	return nil
}

func hookTest(hookfunc NewHookFunc, topic string, t *testing.T) {
	config := nsq.NewConfig()
	client, err := nsq.NewProducer("127.0.0.1:4150", config)
	if err != nil {
		log.Panic(err)
	}
	hook, err := hookfunc(client, "localhost", topic, logrus.DebugLevel)
	if err != nil {
		log.Panic(err)
	}
	logrus.AddHook(hook)

	samples := 100
	for index := 0; index < samples; index++ {
		logrus.Infof("Hustej msg %d", time.Now().Unix())
	}

	// Allow time for data to be processed.
	consumer, err := nsq.NewConsumer(topic, "test", config)
	if err != nil {
		log.Panic(err)
	}
	ch := make(chan int)
	consumer.AddHandler(&nsqHandle{ch})

	consumer.ConnectToNSQD("127.0.0.1:4150")
	defer consumer.Stop()

	tch := time.After(100 * time.Second)

	var count int64
	for {
		select {
		case _, ok := <-ch:
			if ok {
				count += 1
			}
			log.Println("count", count)
			if count == int64(samples) {
				return
			}

		case <-tch:
			t.Errorf("Not all logs pushed to elastic: expected %d got %d", samples, count)
			t.FailNow()
		}
	}
}
