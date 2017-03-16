package producer

import (
	"encoding/json"
	"log"
	"os"
	"time"
	"trial/common"

	"github.com/bitly/go-nsq"
)

// RunProducer starts a simple producer which sends json-encoded messages to
// the topic from configuration file. Internal loop never stops.
func RunProducer(appCfg *common.Config) {
	nsqCfg := nsq.NewConfig()
	w, err := nsq.NewProducer(appCfg.NSQHostPort, nsqCfg)
	if err != nil {
		log.Printf("Failed to run producer; %v", err)
		os.Exit(1)
	}
	defer w.Stop()
	for {
		time.Sleep(time.Millisecond * 100)
		now := time.Now()
		msg := common.AppMsg{
			ID:        now.Unix(),
			Timestamp: now.String(),
		}
		bMsg, err := json.Marshal(msg)
		if err != nil {
			log.Printf("Failed to send message %v; %v", msg, err)
		} else {
			if err := w.Publish(appCfg.TopicName, bMsg); err != nil {
				log.Printf("Failed to write message; %v", err)
			}
		}
	}
}
