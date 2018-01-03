package utilities

/**
 * Copyright 2016 Confluent Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *****************************************************************
 * Modified by Alfonso Elizalde on Dec 30th 2017 for torbit-chat
 *****************************************************************
 */

// consumer_example implements a consumer using the non-channel Poll() API
// to retrieve messages and events.

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"os"
	"torbit/persistence"
	"time"
)

//Keeping new list of messages here
var CurrentMessages string

//Kafka consumer listener
func  KafkaConsumerActiveListener(group string,kafkaTopic string) {

	start := time.Now()
	debugTime := time.Now()

	broker := persistence.KafkaHost
	topics := []string{kafkaTopic}

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":    broker,
		"group.id":             group,
		"session.timeout.ms":   persistence.KafkaSessionTimeout,
		"default.topic.config": kafka.ConfigMap{"auto.offset.reset": persistence.KafkaAutoOffsetReset}})

	defer func() {
		c.Close()
		Debug("Kafka consumer connection closed")
	}()

	if err != nil {
		Error(os.Stderr, "Failed to create consumer: ", err)
		os.Exit(1)
	}

	Info("Created Consumer "+c.String())

	err = c.SubscribeTopics(topics, nil)

	run := true
	Debug("Debug response time after created consumer: "+time.Since(debugTime).String())
	debugTime = time.Now()

	for run == true {
		//Polling messages
		ev := c.Poll(100)
		if ev == nil {
			continue
		}

		switch e := ev.(type) {
		case *kafka.Message:
			//If messages are not coming from a user to ignore then we will add it to the list,
			//otherwise will be ignored.
			timestamp,sender,message:=ReadJson(string(e.Value))
			if IgnoreMessageFromContacts(sender,group) == false {
				CurrentMessages += timestamp +" - "+sender+": "+message + "\n"
				Debug("message: "+CurrentMessages)
			}else{
				Info(group + " ignored message: "+string(e.Value) )
			}

		case kafka.PartitionEOF:
			//If kafka is done sending messager or there are no more messages we complete the loop
			Debug("Debug response time with kafka EOF: "+time.Since(debugTime).String())
			debugTime = time.Now()
			run=false
		case kafka.Error:
			//If kafka finish with error we display the error and complete the loop
			Error(os.Stderr, "%% Error: "+e.String())
			run = false
			Info("Debug response time with kafka error: "+time.Since(start).String())
			debugTime = time.Now()
		default:
			//If kafka ignore an event we display and continue
			Debug("Debug response time with kafka Ignored event: "+e.String()+" at "+time.Since(debugTime).String())
			debugTime = time.Now()
		}
	}
}

//Check for new messages saved into CurrentMessages and clean it up
func CheckForMessages() (string){
	var message string
	if CurrentMessages!=""{
		Debug("CurrentMessages has content")
		message=CurrentMessages
		CurrentMessages=""
	}

	return message
}