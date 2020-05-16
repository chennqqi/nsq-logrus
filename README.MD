# Nsq Hook for [Logrus](https://github.com/sirupsen/logrus) 
<img src="https://travis-ci.org/chennqqi/nsqlogrus.svg?branch=master" />

Like <https://github.com/sohlich/elogrus>, but for [nsq](https://github.com/nsqio/go-nsq)

## Changelog

## Usage

```go
package main

import (
	"github.com/sirupsen/logrus"
	"github.com/chennqqi/nsqlogrus"
	"github.com/nsqio/go-nsq"
)

func main() {
	log := logrus.New()
	config := nsq.NewConfig()
	client, err := nsq.NewProducer("127.0.0.1:4150", config)
	if err != nil {
		log.Panic(err)
	}
	hook, err := nsqlogrus.NewAsyncNsqHook(client, "localhost", "mylog", logrus.DebugLevel)
	if err != nil {
		log.Panic(err)
	}
	log.Hooks.Add(hook)

	log.WithFields(logrus.Fields{
		"name": "joe",
		"age":  42,
	}).Error("Hello world!")
}
```

### Asynchronous hook

```go
	...
	nsqlogrus.NewAsyncNsqHook(client, "localhost", "mylog", logrus.DebugLevel)
	...
```
