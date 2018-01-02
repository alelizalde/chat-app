package utilities

import (
	"github.com/gocql/gocql"
	"torbit/persistence"
	"log"
	"fmt"
)

//Creating cassandra session
func CreateSession() (*gocql.Session) {
	cluster := gocql.NewCluster(persistence.CassandraHost)
	cluster.Keyspace = "torbitchat"
	session, _ := cluster.CreateSession()
	return session
}

//Validating is a user is registered to use the chat room
func ValidateUserId(rec_user_id string) (bool){

	session:=CreateSession()
	defer session.Close()

	var user_id string

	if err := session.Query("SELECT user_id FROM torbitchat.users WHERE user_id = ?",rec_user_id).Consistency(gocql.One).Scan(&user_id); err != nil {
		return false
	}

	Debug("User_ID:"+user_id)

	return true
}

//Searching for a message to ignore coming from an specific user into current user
func IgnoreMessageFromContacts(offset string, user_id string) (bool){

	Debug("offset: "+offset)
	session:=CreateSession()
	defer session.Close()

	var contact_id string

	if err := session.Query("SELECT contact_id FROM torbitchat.offset_tracker " +
		"WHERE offset = ?",offset).Consistency(gocql.One).Scan(&contact_id); err != nil {
		//if the offset still doesn't exist either consumer was faster than the database insert
		//or an error happened while insering on offset tracker, we handle this as proceed sending the message
		return false
	}

	Debug("contact_id: "+contact_id)
	Debug("user_id: "+user_id)

	if err := session.Query("SELECT contact_id FROM torbitchat.ignore_contacts " +
		"WHERE contact_id = ? AND user_id = ?",contact_id,user_id).Consistency(gocql.One).Scan(&contact_id); err != nil {
		//current user and contact are not present on the ignore relationship table
		return false
	}

	//we found the current message should be ignored and return true
	return true

}

//Inserting offset number and user sending message into an offset tracker
func InsertOffset(offset int,contact_id string){
	session:=CreateSession()
	defer session.Close()

	fmt.Printf("offset %d contact id %s ",offset,contact_id)

	if err := session.Query(`INSERT INTO offset_tracker (offset,contact_id) VALUES (?, ?)`,
		offset, contact_id).Exec(); err != nil {
		log.Fatal(err)
	}
}