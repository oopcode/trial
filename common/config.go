package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
)

const (
	cConfigPath = "/opt/trial/config.json"
)

// Config holds application configuration.
type Config struct {
	TopicName          string `json:"nsq_topic_name"`
	NSQHostPort        string `json:"nsq_host_port"`
	NSQConsumerDelta   int    `json:"nsq_consumer_delta"`
	NSQConsumerMaxRead int    `json:"nsq_consumer_max_read"`
	ASHost             string `json:"as_host"`
	ASPort             int    `json:"as_port"`
	ASNamespace        string `json:"as_namespace"`
	ASSet              string `json:"as_set"`
}

// NewConfig is a constructor for Config; it also loads config params.
func NewConfig() (*Config, error) {
	if _, err := os.Stat(cConfigPath); err != nil {
		log.Printf("Configuration file %s not found", cConfigPath)
		return nil, err
	}
	log.Println("Found config, reading...")
	var (
		cfg           = &Config{}
		configFile, _ = os.Open(cConfigPath)
		jsonParser    = json.NewDecoder(configFile)
	)
	if err := jsonParser.Decode(cfg); err != nil {
		log.Printf("Failed to read config; %v", err)
		return nil, err
	}
	if err := cfg.Check(); err != nil {
		log.Printf("Invalid configuration; %v", err)
		return nil, err
	}
	log.Printf("Successfully read config; %+v", cfg)
	return cfg, nil
}

// Check performs sanity checks for config params.
func (c *Config) Check() error {
	if err := c.checkHost(c.ASHost); err != nil {
		return err
	}
	if err := c.checkHostPort(c.NSQHostPort); err != nil {
		return err
	}
	if c.NSQConsumerDelta < 1 {
		return fmt.Errorf("ConsumerDelta should be greater than 0, got %d",
			c.NSQConsumerDelta)
	}
	if c.NSQConsumerMaxRead < 1 {
		return fmt.Errorf("ConsumerMaxRead should be greater than 0, got %d",
			c.NSQConsumerDelta)
	}
	if len(c.ASNamespace) < 1 {
		return errors.New("Aerospike namespace name can't be an empty string")
	}
	return nil
}

// checkHost is a mock for checking a host string for validity.
func (c *Config) checkHost(host string) error {
	// TODO: implement.
	return nil
}

// checkHostPort is a mock for checking a host:port string for validity.
func (c *Config) checkHostPort(hostPort string) error {
	// TODO: implement.
	return nil
}
