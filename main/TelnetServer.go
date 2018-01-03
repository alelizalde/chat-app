package main

import (

	"net"
	"bufio"
	"strings"
	"torbit/utilities"
	"torbit/persistence"
	"time"
	"log"
	"os"
	"io"
	"bytes"
)

//Creating interface for jobRunner to implement run
type JobInterface interface {
	run() string
}


type Job        struct { param string }
type chatMessage   struct { Job }
type InvalidJob struct { Job }
var RemoteIP string
var KafkaTopic string

//Executing run for ChatMessage
func (job chatMessage) run() string {
	//Sending message to kafka
	utilities.KafkaProducer(time.Now().Format("2006.01.02 15:04:05"),"telnet", "broadcast message "+ strings.TrimSuffix(job.param,"\r"),KafkaTopic)
	//Saving message into log file
	utilities.Savelog(time.Now().Format("2006.01.02 15:04:05") + " - telnet broadcast to "+KafkaTopic+" room from "+RemoteIP+": "+ job.param)
	return "Message sent to chat room " + KafkaTopic + ": "+ job.param
}

//Executing run for InvalidJob
func (job InvalidJob) run() string {
	return "Invalid message"
}

//job runner function for request handler
func jobRunner(job JobInterface, out chan string) {
	utilities.Block{
		Try: func() {
			out <- job.run() + "\n"
		},
		Catch: func(e utilities.Exception) {
			utilities.Error(e)
		},
		Finally: func() {
			utilities.Info("Request completed...")
		},
	}.Do()
}

//Decision maker factory
func jobFactory(message string) JobInterface {
	utilities.Debug("message: "+message)
	if strings.Trim(message," ") != "" {
		return chatMessage{Job{message}}
	}
	return InvalidJob{Job{""}}
}

//Request handler
func requestHandler(conn net.Conn, out chan string) {
	defer close(out)

	for {
		line, err := bufio.NewReader(conn).ReadBytes('\n')
		if err != nil { return }

		job := jobFactory(strings.TrimRight(string(line), "\n"))
		go jobRunner(job, out)
	}
}

//Send data from telnet client
func sendData(conn net.Conn, in <-chan string) {
	defer conn.Close()

	for {
		message := <- in
		if message==""{ break }
		utilities.Debug(message)
		io.Copy(conn, bytes.NewBufferString(message))
	}
}

func main() {

	//Validating we have enough parameters
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

	//Setting up global variables
	persistence.Logfile=cfg.String("logfile")
	persistence.KafkaHost=cfg.String("kafka-host")
	KafkaTopic=os.Args[2]
	if KafkaTopic==""{
		utilities.Error("Group name cannot be empty")
		os.Exit(1)
	}

	persistence.KafkaSessionTimeout=cfg.String("kafka.session.timeout.ms")
	persistence.KafkaAutoOffsetReset=cfg.String("kafka.auto.offset.reset")
	persistence.CassandraHost=cfg.String("cassandra.host")
	utilities.Debug("Kafka auto offset: "+persistence.CassandraHost)

	utilities.Info("Chat room: "+KafkaTopic)

	//Openning Listener in specified port
	psock, err := net.Listen("tcp", ":"+cfg.String("port"))

	if err != nil {log.Fatal(err)}

	//Listening to telnet requests
	for {
		conn, err := psock.Accept()
		if err != nil { return }
		RemoteIP=conn.RemoteAddr().String()
		utilities.Debug("Remote IP: "+RemoteIP)
		utilities.Debug("UserId: "+strings.Split(conn.RemoteAddr().String(),":")[0])
		channel := make(chan string)
		go requestHandler(conn, channel)
		go sendData(conn, channel)
	}
}