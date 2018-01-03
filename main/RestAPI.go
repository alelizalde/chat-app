package main

import (
	"log"
	"net/http"
	"github.com/gorilla/mux"
	"time"
	"fmt"
	"html"
	"torbit/utilities"
	"os"
	"torbit/persistence"
	"io/ioutil"
	"io"
)

func main() {

	//Validating we have enough parameters
	if len(os.Args)<2{
		utilities.Error("You need to provide the path for the config file as parameter, please go to README.md.")
		os.Exit(1)
	}

	//Reading configuration file
	cfg,err := utilities.LoadConfig(os.Args[1])

	if err != nil {
		panic(err)
	}

	//setting global variables
	persistence.Logfile=cfg.String("logfile")
	utilities.Debug("Log file: "+persistence.Logfile)
	persistence.KafkaHost=cfg.String("kafka-host")
	utilities.Debug("Host: "+persistence.KafkaHost)
	persistence.KafkaSessionTimeout=cfg.String("kafka.session.timeout.ms")
	utilities.Debug("Kafka timeout: "+persistence.KafkaSessionTimeout)
	persistence.KafkaAutoOffsetReset=cfg.String("kafka.auto.offset.reset")
	utilities.Debug("Kafka auto offset: "+persistence.KafkaAutoOffsetReset)
	persistence.CassandraHost=cfg.String("cassandra.host")
	utilities.Debug("Cassandra host: "+persistence.CassandraHost)

	//routing to different functions
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", HomeHandler).Methods("GET")
	router.HandleFunc("/send/{room-name}/{userid}", SendMessage).Methods("POST")
	router.HandleFunc("/receive/{room-name}/{userid}", GetMessage).Methods("GET")
	http.Handle("/", router)

	utilities.Info("Server is up")

	//listening in port 8080
	log.Fatal(http.ListenAndServe(":8080", nil))


}

//Handling request for root
func HomeHandler(w http.ResponseWriter, r *http.Request){

	start := time.Now()
	fmt.Fprintf(w, "Torbit chat, %q", html.EscapeString(r.URL.Path))

	log.Printf(
		"%s\t%s\t%s",
		r.Method,
		r.RequestURI,
		time.Since(start),
	)

	utilities.Info("response time: "+time.Since(start).String())

}

//Send messages from one user to a chat room
func SendMessage(w http.ResponseWriter, r *http.Request){

	//Starting time to monitor response times
	start := time.Now()

	//reading user_id and chat room from path
	vars := mux.Vars(r)

	//Validating user id is not empty
	if utilities.ValidateUserId(vars["userid"]) == false {
		utilities.Info("Invalid user id")
		fmt.Fprintf(w, "Invalid user id")
	} else {
		utilities.Debug("UserId: "+vars["userid"])
	}

	//Validating chat room is not empty
	utilities.Debug("Chat room name: "+vars["room-name"])
	if vars["room-name"] == ""{
		utilities.Debug("Chat group name is empty")
		fmt.Fprintf(w, "Chat group name cannot be empty")
	}

	//if user id and chat room are not empty we can send the message
	if vars["userid"]!= "" && vars["room-name"] != "" {

		//limiting the message to 1048576 bytes
		body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))

		if err != nil {
			panic(err)
		}
		if err := r.Body.Close(); err != nil {
			panic(err)
		}

		//Sending message to kafka
		utilities.KafkaProducer(time.Now().Format("2006.01.02 15:04:05"), vars["userid"], string(body),vars["room-name"])
		//Saving message to log file
		utilities.Savelog(time.Now().Format("2006.01.02 15:04:05") + " - " + vars["userid"] + " sent to " + vars["room-name"] + " room from " + r.RemoteAddr + " :" + string(body))
	}

	//Printing response time
	utilities.Info("response time: "+time.Since(start).String())

}

//Receive pending messages sent to a room and user id
func GetMessage(w http.ResponseWriter, r *http.Request) {
	//Starting time to monitor response times
	start := time.Now()
	//Starting time to monitor response times on debugging points
	debugTime := time.Now()

	//reading user_id and chat room from path
	vars := mux.Vars(r)

	//Validating user id is not empty
	if utilities.ValidateUserId(vars["userid"]) == false {
		utilities.Info("Invalid user id")
		fmt.Fprintf(w, "Invalid user id")
	} else {
		utilities.Debug("UserId: " + vars["userid"])
	}

	//Validating chat room is not empty
	utilities.Debug("Chat room name: " + vars["room-name"])
	if vars["room-name"] == "" {
		utilities.Debug("Chat group name is empty")
		fmt.Fprintf(w, "Chat group name cannot be empty")
	}


	utilities.Debug("Debug response time set config: " + time.Since(debugTime).String())
	debugTime = time.Now()

	//if user id and chat room are not empty we can receive the message
	if vars["userid"] != "" && vars["room-name"] != "" {

		//Receiving message from Kafka
		utilities.KafkaConsumerActiveListener(vars["userid"],vars["room-name"])

		//Sending response back
		fmt.Fprintf(w, "%s", utilities.CheckForMessages())

		utilities.Debug("Debug response time after sending response message: " + time.Since(debugTime).String())
		debugTime = time.Now()
	}

	utilities.Debug("Response completed")
	utilities.Info("response time: "+time.Since(start).String())
}