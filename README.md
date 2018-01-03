Torbit chat
==============
Programming challenge for Torbit/Walmartlabs to Alfonso Elizalde
--------------

Details
--------------
I decided to use kafka to store and handle messages on the chat using kafka topics as chat rooms and topic consumer groups as user ids logged into a chat room, this is giving me the ability to send the messages to all users subscribed/listening to a chat room. I also implemented a cassandra cluster to store metadata including valid/registered users, ignored contacts when a user does not want to receive messages from a specific sender. 

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
    │   ├── MessageAnalyzer.go
    │   ├── RestAPI.go  
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
3. MessageAnalyzer.go, implemented to increase the count of a word when is found in an message, this implementation is decoupled from the dest of the chat operation as I didn't want to slow down the chat message consumer and producer, also the chat is always listening for new messages and the rest API listener only search for new messages by request. It needs to receive the configuration file path and name  and chat room name as parameters. 
    ```Shell
    go run main/MessageAnalyzer.go ./torbitchat.conf torbit-room-1
    ```
    Query the database to monitor popularity:  
    ```sql
    cqlsh:torbitchat> select * from torbitchat.popular_count;
    ```


References
--------------
* I used (https://parroty00.wordpress.com/2013/07/18/golang-tcp-server-example/) as example for telnet implementation.

* I used (https://astaxie.gitbooks.io/build-web-application-with-golang/en/13.4.html) as example to handle configuration and logging.

* Used different websites to create the rest service.

Dependencies
--------------
* Using confluentinc/confluent-kafka-go: https://github.com/confluentinc/confluent-kafka-go to implement kafka producer and consumer, these libraries have dependency with librdkafka, if you have a mac you can follow https://github.com/confluentinc/confluent-kafka-python/issues/6 to install librdkafka.
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
    
    CREATE TABLE torbitchat.ignore_contacts (
        contact_id text,
        user_id text,
        PRIMARY KEY (contact_id, user_id)
    );
    
    CREATE TABLE torbitchat.words_analytics (
    word text PRIMARY KEY
    );
    
    CREATE TABLE popular_count (
    word text PRIMARY KEY,
    popularity counter
    );
    ```

Pending activities / known issues
-----------------
* Log file needs to be purged periodically, currently there is only one file growing with all the messages.
* Security needs to be implemented into the chat room:
    * User and password
    * TLS over http and telnet
    * Need to have implementation for chat rooms administration, right now any user registered on the database can join any chat room as there is no concept of private rooms implemented.
* Kafka messages are set to be purged every 24 hours, this is a kafka configuration and can be changed.

Additional implementation
-----------------
* We can capture and handle some words coming from the messages for different purposes - I implemented this feature to count popularity.
* The chat is limited to 1000 bytes message size to prevent big messages getting pushed.
