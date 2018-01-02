Torbit chat
==============
Programming challenge for Torbit/Walmartlabs to Alfonso Elizalde
--------------

Details
--------------
I decided to use kafka to store and handle messages on the chat using kafka topics as chat rooms and topic consumer groups as user ids logged into a chat room, this is giving me the ability to send the messages to all users subscribed/listening to a chat room. I also implemented a cassandra cluster to store metadata including valid/registered users, ignored contacts and a kafka offset tracker to registry who send what messages and ignore them when a user does not what to receive messages from a specific contact. 

Project anatomy
--------------
Project name: torbit

* 3 packages:  
    * main - Main package  
    * persistence - To store global persistence values  
    * utilities - all common utilities

    ```text
    .torbit  
    ├── README.md  
    ├── main  
    │   ├── RestAPI.go  
    │   └── TelnetServer.go  
    ├── persistence  
    │   └── GlobalPersistance.go  
    ├── torbitchat.conf  
    └── utilities  
        ├── DBHandler.go  
        ├── ErrorHandler.go  
        ├── KafkaConsumer.go  
        ├── KafkaProducer.go  
        └── Logging.go 
    ``` 

     


Usage
--------------
There are two main programs:
1. RestAPI.go to start the rest service. It needs to receive the configuration file path and name as parameter.  
    ```Shell
     go run main/RestAPI.go ./torbitchat.conf 
    ```
    
    To send messages using the rest service:  
    ```text
    http://server:8080/send/{chat room}/{user id} - And message needs to be in the body of the request
    i.e. http://192.168.0.16:8080/send/torbit-room-1/alfonso
    ```   
       
    ```text
    http://server:8080/receive/{chat room}/{user id}
    http://192.168.0.16:8080/receive/torbit-room-1/wen  
    ```
2. TelnetServer.go to start telnet server - It needs to receive the configuration file path and name  and chat room name as parameters.  
    ```Shell
    go run main/TelnetServer.go ./torbitchat.conf torbit-room-1
    ```
  
    To open a telnet session:  
    ```text
    telnet host port
    i.e. telnet localhost 5000
    ```



References
--------------
* I used (https://parroty00.wordpress.com/2013/07/18/golang-tcp-server-example/) as example for telnet implementation.

* I used (https://astaxie.gitbooks.io/build-web-application-with-golang/en/13.4.html) as example to handle configuration and logging.

Dependencies
--------------
* Using confluentinc/confluent-kafka-go: https://github.com/confluentinc/confluent-kafka-go to implement kafka producer and consumer, these libraries have dependency with librdkafka.
* Using gocql/gocql: https://github.com/gocql/gocql to connect to cassandra cluster.
* Usinf gorilla/mux: https://github.com/gorilla/mux to handle rest api routing.

External implementations:
--------------
* kafka 2.10-0.10.1.1 to store, receive and broadcast chat messages
    
* Cassandra 3.11.1 to store metadata
    
    ```sql
    Schema: 
    
    CREATE KEYSPACE torbitchat;
    
    CREATE TABLE torbitchat.users (
        user_id text PRIMARY KEY,
        name text
    );
    
    CREATE TABLE torbitchat.offset_tracker (
        offset int PRIMARY KEY,
        contact_id text
    );
    
    CREATE TABLE torbitchat.ignore_contacts (
        contact_id text,
        user_id text,
        PRIMARY KEY (contact_id, user_id)
    );
    ```

Pending activities / known issues
-----------------
* Log file needs to be purged periodically, currently there is only one file growing with all the messages.
* Kafka listener could be faster than the insert into cassandra offset track used to check if there are messages from users to ignore, if this happen it is assumed the message is not ignored. I need to implement a retry method when offset doesn't exist yet in cassandra offset_tracker table and kafka consumer alrealy has the new message.
* Security needs to be implemented into the chat room:
    * User and password
    * TLS over http and telnet
    * We might want to capture and handle curse words (optional)
    * Need to have implementation for chat rooms administration, right now any user registered on the database can join any chat room as there is no concept of private rooms implemented.
* Kafka messages are set to be purged every 24 hours, this is a kafka configuration and can be changed.
