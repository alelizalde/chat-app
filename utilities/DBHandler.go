package utilities

import (
	"github.com/gocql/gocql"
	"torbit/persistence"
)

//Creating cassandra session
func CreateSession() (*gocql.Session) {
	Debug("Connecting to cassandra")
	cluster := gocql.NewCluster(persistence.CassandraHost)
	cluster.Keyspace = "torbitchat"
	session, _ := cluster.CreateSession()
	Debug("Cassandra session created")
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
func IgnoreMessageFromContacts(contact_id string, user_id string) (bool){

	session:=CreateSession()
	defer session.Close()

	Debug("contact_id: "+contact_id)
	Debug("user_id: "+user_id)

	if err := session.Query("SELECT contact_id FROM torbitchat.ignore_contacts " +
		"WHERE contact_id = ? AND user_id = ?",contact_id,user_id).Consistency(gocql.One).Scan(&contact_id); err != nil {
		//current user and contact are not present on the ignore relationship table
		return false
	}

	//we found the current message should be ignored
	return true

}