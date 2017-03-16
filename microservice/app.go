package microservice

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
	"trial/common"

	as "github.com/aerospike/aerospike-client-go"
	"github.com/bitly/go-nsq"
)

// App implements a microservice which gets data from NSQ and sends it to
// Aerospike.
type App struct {
	sync.Mutex
	conn                *nsq.Consumer
	cfg                 *common.Config
	readCount           int
	consumptionInterval time.Duration
	resumeCh            chan struct{}
}

// NewApp is a constructor for Consumer.
func NewApp(cfg *common.Config) *App {
	return &App{
		cfg:                 cfg,
		consumptionInterval: time.Duration(cfg.NSQConsumerDelta) * time.Second,
		resumeCh:            make(chan struct{}),
	}
}

// Run starts consumption from NSQ.
func (a *App) Run() error {
	nsqCfg := nsq.NewConfig()
	conn, err := nsq.NewConsumer(a.cfg.TopicName, "ch", nsqCfg)
	if err != nil {
		return fmt.Errorf("Failed to create a consumer; %v", err)
	}
	go a.loop()
	conn.AddHandler(a)
	if err := conn.ConnectToNSQD(a.cfg.NSQHostPort); err != nil {
		return fmt.Errorf("Failed to connect to NSQ server; %v", err)
	}
	a.conn = conn
	return nil
}

// Kill closes connection to NSQ.
func (a *App) Kill() {
	if err := a.conn.DisconnectFromNSQD(a.cfg.NSQHostPort); err != nil {
		log.Printf("Failed to disconnect from NSQ; %v", err)
	}
}

// HandleMessage satisfies nsq.Handler interface. All production logic is
// implemented in handleMessage().
func (a *App) HandleMessage(msg *nsq.Message) error {
	if a.canHandle() {
		return a.handleMessage(msg)
	}
	<-a.resumeCh
	return a.handleMessage(msg)
}

func (a *App) handleMessage(msg *nsq.Message) error {
	a.incReadCount()
	appMsg := &common.AppMsg{}
	if err := json.Unmarshal(msg.Body, appMsg); err != nil {
		log.Printf("Failed to read message; %s, %v", string(msg.Body), err)
		return err
	}
	if err := appMsg.Check(); err != nil {
		log.Printf("Corrupt message; %+v, %v", appMsg, err)
		return err
	}
	log.Printf("Received message: %+v", appMsg)
	a.sendMessageAS(appMsg)
	return nil
}

// sendMessageAS writes the message to aerospike. A new connection is created
// for each write (inefficient, but spares some code required to maintain a
// living connection).
func (a *App) sendMessageAS(msg *common.AppMsg) {
	client, err := as.NewClient(a.cfg.ASHost, a.cfg.ASPort)
	if err != nil {
		log.Printf("Failed to connect to aerospike; %v", err)
		return
	}
	key, err := as.NewKey(a.cfg.ASNamespace, a.cfg.ASSet, msg.ID)
	if err != nil {
		log.Printf("Failed to create aerospike key; %v", err)
		return
	}
	tsBin := as.NewBin("timestamp", msg.Timestamp)
	err = client.PutBins(nil, key, tsBin)
	if err != nil {
		log.Printf("Failed to put aerospike bins; %v", err)
		return
	}
	client.Close()
}

func (a *App) loop() {
	ticker := time.NewTicker(a.consumptionInterval)
	for {
		select {
		case <-ticker.C:
			a.rstReadCount()
			a.resumeCh <- struct{}{}
		}
	}
}

func (a *App) canHandle() bool {
	a.Lock()
	defer a.Unlock()
	return a.readCount < a.cfg.NSQConsumerMaxRead
}

func (a *App) incReadCount() {
	a.Lock()
	defer a.Unlock()
	a.readCount++
}

func (a *App) rstReadCount() {
	a.Lock()
	defer a.Unlock()
	a.readCount = 0
}
