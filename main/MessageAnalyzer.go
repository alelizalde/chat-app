package main

import (
	"strings"
	"torbit/utilities"
	"torbit/persistence"
	"os"
	"log"
	"time"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)


func main() {

	if len(os.Args)<3{
		utilities.Error("You need to provide the path for the config file and chat room name to connect as parameters, please go to README.md.")
		os.Exit(1)
	}

	//Reading configuration file
	cfg,err := utilities.LoadConfig(os.Args[1])
	if cfg == nil {
		utilities.Error("Configuration file not found")
		os.Exit(1)
	}

	if err != nil {log.Fatal(err)}

	//Setting up global variables
	persistence.Logfile=cfg.String("logfile")
	persistence.KafkaHost=cfg.String("kafka-host")
	if os.Args[2]==""{
		utilities.Error("Chat room name cannot be empty")
		os.Exit(1)
	}

	persistence.KafkaSessionTimeout=cfg.String("kafka.session.timeout.ms")
	persistence.KafkaAutoOffsetReset=cfg.String("kafka.auto.offset.reset")
	persistence.CassandraHost=cfg.String("cassandra.host")
	utilities.Debug("Kafka auto offset: "+persistence.CassandraHost)

	utilities.Info("Chat room: "+os.Args[2])


	utilities.LoadWordsForAnalysis()

	start := time.Now()
	debugTime := time.Now()

	broker := persistence.KafkaHost
	topics := []string{os.Args[2]}

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":    broker,
		"group.id":             "torbit-analytics",
		"session.timeout.ms":   persistence.KafkaSessionTimeout,
		"default.topic.config": kafka.ConfigMap{"auto.offset.reset": persistence.KafkaAutoOffsetReset}})

	defer func() {
		c.Close()
		utilities.Debug("Kafka consumer connection closed")
	}()

	if err != nil {
		utilities.Error(os.Stderr, "Failed to create consumer: ", err)
		os.Exit(1)
	}

	utilities.Info("Created Consumer "+c.String())

	err = c.SubscribeTopics(topics, nil)

	run := true
	utilities.Debug("Debug response time after created consumer: "+time.Since(debugTime).String())
	debugTime = time.Now()

	for run == true {
		//Polling messages
		ev := c.Poll(100)
		if ev == nil {
			continue
		}

		switch e := ev.(type) {
		case *kafka.Message:
			_,_,message:=utilities.ReadJson(string(e.Value))
			for _, word := range persistence.WordsToaAnalyze{
				if strings.Contains(message,word)==true{
					utilities.Info("Counter incremented for word: " + word)
					utilities.IncrementCount(word)
				}
			}
		case kafka.PartitionEOF:
			//If kafka is done sending messager or there are no more messages we complete the loop
			utilities.Debug("Debug response time with kafka EOF: "+time.Since(debugTime).String())
			debugTime = time.Now()
		case kafka.Error:
			//If kafka finish with error we display the error and complete the loop
			utilities.Error(os.Stderr, "%% Error: "+e.String())
			run = false
			utilities.Info("Debug response time with kafka error: "+time.Since(start).String())
			debugTime = time.Now()
		default:
			//If kafka ignore an event we display and continue
			utilities.Debug("Debug response time with kafka Ignored event: "+e.String()+" at "+time.Since(debugTime).String())
			debugTime = time.Now()
		}
	}

}