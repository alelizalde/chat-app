package utilities

import (
	"os"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"torbit/persistence"
)

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


 //Kafka producer
func KafkaProducer(message string,user_id string,kafkaTopic string) {

	broker := persistence.KafkaHost
	topic := kafkaTopic

	//Creating new producer
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": broker})

	if err != nil {
		Error(err)
		os.Exit(1)
	}

	Info("Created Producer "+p.String()+"\n")

	// Optional delivery channel, if not specified the Producer object's
	// .Events channel is used.
	deliveryChan := make(chan kafka.Event)

	value := message

	//Sending message
	err = p.Produce(&kafka.Message{TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny}, Value: []byte(value)}, deliveryChan)

	e := <-deliveryChan
	m := e.(*kafka.Message)

	if m.TopicPartition.Error != nil {
		Error(m.TopicPartition.Error)
	} else {
		Info("Delivered message to topic "+string(*m.TopicPartition.Topic)+" at offset "+m.TopicPartition.Offset.String()+"\n")
		//Inserting new message into the database
		InsertOffset(int(m.TopicPartition.Offset),user_id)
	}

	close(deliveryChan)
}