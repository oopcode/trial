package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"trial/common"
	"trial/microservice"
	"trial/producer"
)

func main() {
	common.WritePid()
	common.SetupLog()
	cfg, err := common.NewConfig()
	if err != nil {
		log.Println("Failed to run application, exiting")
		os.Exit(1)
	}
	go producer.RunProducer(cfg)
	log.Println("Started producer")
	app := microservice.NewApp(cfg)
	if err := app.Run(); err != nil {
		log.Printf("Failed to run app; %v", err)
		os.Exit(1)
	}
	go handleSignals(app)
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}

func handleSignals(killable common.IKillable) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigs
	log.Printf("Caught signal %v; exiting gracefully", sig.String())
	killable.Kill()
	log.Println("Exited")
	os.Exit(1)
}
