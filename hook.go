package nsqlogrus

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/sirupsen/logrus"
)

type fireFunc func(entry *logrus.Entry, hook *NsqHook) error

// NsqHook is a logrus
// hook for nsq
type NsqHook struct {
	client    *nsq.Producer
	topic     string
	host      string
	levels    []logrus.Level
	ctx       context.Context
	ctxCancel context.CancelFunc
	fireFunc  fireFunc
}

type message struct {
	Host      string `json:"Host,omitempty"`
	Timestamp string `json:"@timestamp"`
	File      string `json:"File,omitempty"`
	Func      string `json:"Func,omitempty"`
	Message   string `json:"Message,omitempty"`
	Data      logrus.Fields
	Level     string `json:"Level,omitempty"`
}

// NewNsqHook creates new hook.
// client - nsq producer client
// host - host of system
// topic - name of nsqd
// level - log level
func NewNsqHook(client *nsq.Producer, host, topic string, level logrus.Level) (*NsqHook, error) {
	return newHookFuncAndFireFunc(client, host, topic, level, syncFireFunc)
}

// NewAsyncNsqHook creates new  hook with asynchronous log.
// client - nsq producer client
// host - host of system
// topic - name of nsqd
// level - log level
// index - name of the index in ElasticSearch
func NewAsyncNsqHook(client *nsq.Producer, host, topic string, level logrus.Level) (*NsqHook, error) {
	return newHookFuncAndFireFunc(client, host, topic, level, asyncFireFunc)
}

func newHookFuncAndFireFunc(client *nsq.Producer, host, topic string, level logrus.Level, fireFunc fireFunc) (*NsqHook, error) {
	var levels []logrus.Level
	for _, l := range []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
		logrus.TraceLevel,
	} {
		if l <= level {
			levels = append(levels, l)
		}
	}

	ctx, cancel := context.WithCancel(context.TODO())
	return &NsqHook{
		client:    client,
		host:      host,
		topic:     topic,
		levels:    levels,
		ctx:       ctx,
		ctxCancel: cancel,
		fireFunc:  fireFunc,
	}, nil
}

// Fire is required to implement
// Logrus hook
func (hook *NsqHook) Fire(entry *logrus.Entry) error {
	return hook.fireFunc(entry, hook)
}

func asyncFireFunc(entry *logrus.Entry, hook *NsqHook) error {
	client := hook.client
	msg := createMessage(entry, hook)
	txt, _ := json.Marshal(msg)
	return client.Publish(hook.topic, txt)
}

func createMessage(entry *logrus.Entry, hook *NsqHook) *message {
	level := entry.Level.String()

	if e, ok := entry.Data[logrus.ErrorKey]; ok && e != nil {
		if err, ok := e.(error); ok {
			entry.Data[logrus.ErrorKey] = err.Error()
		}
	}

	var file string
	var function string
	if entry.HasCaller() {
		file = entry.Caller.File
		function = entry.Caller.Function
	}

	return &message{
		hook.host,
		entry.Time.UTC().Format(time.RFC3339Nano),
		file,
		function,
		entry.Message,
		entry.Data,
		strings.ToUpper(level),
	}
}

func syncFireFunc(entry *logrus.Entry, hook *NsqHook) error {
	client := hook.client
	msg := createMessage(entry, hook)
	txt, _ := json.Marshal(msg)
	return client.PublishAsync(hook.topic, txt, nil)
}

// Levels Required for logrus hook implementation
func (hook *NsqHook) Levels() []logrus.Level {
	return hook.levels
}

// Cancel all calls to elastic
func (hook *NsqHook) Cancel() {
	hook.ctxCancel()
}
