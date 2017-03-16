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
// Aerospike. It also implements custom consumption procedure: only
// Config.NSQConsumerMaxRead messages are read each Config.NSQConsumerDelta
// seconds.
type App struct {
	sync.Mutex
	nsqConn     *nsq.Consumer
	asConn      *common.ASConn
	cfg         *common.Config
	msg2delayed map[*nsq.Message]chan error
	stopper     chan struct{}
}

// NewApp is a constructor for Consumer.
func NewApp(cfg *common.Config) *App {
	return &App{
		cfg:         cfg,
		msg2delayed: make(map[*nsq.Message]chan error),
		stopper:     make(chan struct{}),
	}
}

// Run starts consumption from NSQ.
func (a *App) Run() error {
	nsqCfg := nsq.NewConfig()
	nsqCfg.Set("max_in_flight", a.cfg.NSQConsumerMaxRead)
	nsqConn, err := nsq.NewConsumer(a.cfg.TopicName, "ch", nsqCfg)
	if err != nil {
		return fmt.Errorf("Failed to create a consumer; %v", err)
	}
	go a.Start()
	// Set up NSQ comsumer.
	nsqConn.AddConcurrentHandlers(a, a.cfg.NSQConsumerMaxRead)
	if err := nsqConn.ConnectToNSQD(a.cfg.NSQHostPort); err != nil {
		return fmt.Errorf("Failed to connect to NSQ server; %v", err)
	}
	a.nsqConn = nsqConn
	// Set up connection to Aerospike.
	a.asConn = common.NewASConn(a.cfg.ASHost, a.cfg.ASPort)
	return nil
}

// HandleMessage satisfies nsq.Handler interface.
func (a *App) HandleMessage(msg *nsq.Message) error {
	delayed := a.scheduleMessage(msg)
	// Waiting for delayed output will block NSQ's processing routine; as we
	// have set "max_in_flight" ti NSQConsumerMaxRead, no new messages will
	// be read until delayed messages are processed  with an execute() call.
	return <-delayed
}

// Start initializes app's processing loop.
func (a *App) Start() {
	ticker := time.NewTicker(
		time.Duration(a.cfg.NSQConsumerDelta) * time.Second)
	for {
		select {
		case <-a.stopper:
			return
		case <-ticker.C:
			a.execute()
		}
	}
}

// Kill closes connection to NSQ.
func (a *App) Kill() {
	a.Lock()
	defer a.Unlock()
	if err := a.nsqConn.DisconnectFromNSQD(a.cfg.NSQHostPort); err != nil {
		log.Printf("Failed to disconnect from NSQ; %v", err)
	}
	log.Println("Disconnected from NSQ")
	a.asConn.Close()
	log.Println("Disconnected from Aerospike")
	a.stopper <- struct{}{}
	log.Println("Killed app's processing loop")
}

// scheduleMessage saves the message in a.msg2delayed.
func (a *App) scheduleMessage(msg *nsq.Message) chan error {
	a.Lock()
	defer a.Unlock()
	a.msg2delayed[msg] = make(chan error)
	return a.msg2delayed[msg]
}

// execute handles all delayed messages.
func (a *App) execute() {
	a.Lock()
	defer a.Unlock()
	var wg sync.WaitGroup
	for msg, out := range a.msg2delayed {
		wg.Add(1)
		go a.handleMessage(msg, out, &wg)
	}
	wg.Wait()
	a.msg2delayed = make(map[*nsq.Message]chan error)
}

func (a *App) handleMessage(msg *nsq.Message, delayed chan error,
	wg *sync.WaitGroup) {
	defer wg.Done()
	appMsg := &common.AppMsg{}
	if err := json.Unmarshal(msg.Body, appMsg); err != nil {
		log.Printf("Failed to read message; %s, %v", string(msg.Body), err)
		delayed <- err
		return
	}
	if err := appMsg.Check(); err != nil {
		log.Printf("Corrupt message; %+v, %v", appMsg, err)
		delayed <- err
		return
	}
	log.Printf("Received message: %+v", appMsg)
	if err := a.sendMessageAS(appMsg); err != nil {
		log.Printf("Failed to send to aerospike; %v", err)
		delayed <- err
		return
	}
	delayed <- nil
}

// sendMessageAS writes the message to aerospike.
func (a *App) sendMessageAS(msg *common.AppMsg) error {
	key, err := as.NewKey(a.cfg.ASNamespace, a.cfg.ASSet, msg.ID)
	if err != nil {
		log.Printf("Failed to create aerospike key; %v", err)
		return err
	}
	tsBin := as.NewBin("timestamp", msg.Timestamp)
	err = a.asConn.PutBins(nil, key, tsBin)
	if err != nil {
		log.Printf("Failed to put aerospike bins; %v", err)
		return err
	}
	return nil
}
