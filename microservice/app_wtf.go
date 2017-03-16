package microservice

import (
	"fmt"
	"log"
	"time"
	"trial/common"

	"github.com/bitly/go-nsq"
)

// AppWTF implements a microservice which gets data from NSQ and sends it to
// Aerospike.
type AppWTF struct {
	conn *nsq.Consumer
	cfg  *common.Config
}

// NewAppWTF is a constructor for Consumer.
func NewAppWTF(cfg *common.Config) *AppWTF {
	return &AppWTF{
		cfg: cfg,
	}
}

// Run starts consumption from NSQ.
func (a *AppWTF) Run() error {
	nsqCfg := nsq.NewConfig()
	conn, err := nsq.NewConsumer(a.cfg.TopicName, "ch", nsqCfg)
	if err != nil {
		return fmt.Errorf("Failed to create a consumer; %v", err)
	}
	// N.B.: LOTS of concurrent handlers.
	conn.AddConcurrentHandlers(a, 100)
	if err := conn.ConnectToNSQD(a.cfg.NSQHostPort); err != nil {
		return fmt.Errorf("Failed to connect to NSQ server; %v", err)
	}
	a.conn = conn
	return nil
}

// HandleMessage satisfies nsq.Handler interface.
func (a *AppWTF) HandleMessage(msg *nsq.Message) error {
	time.Sleep(time.Second)
	return nil
}

// Kill closes connection to NSQ.
func (a *AppWTF) Kill() {
	if err := a.conn.DisconnectFromNSQD(a.cfg.NSQHostPort); err != nil {
		log.Printf("Failed to disconnect from NSQ; %v", err)
	}
}
