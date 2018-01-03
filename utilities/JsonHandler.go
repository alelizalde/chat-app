package utilities

import (
	"encoding/json"
	"time"
)

//Read incoming json message
func ReadJson(message string) (string,string,string) {

	byt := []byte(message)
	var dat map[string]interface{}

	if err := json.Unmarshal(byt, &dat); err != nil {
		return time.Now().Format("2006.01.02 15:04:05"),"error","error reading message"
	}

	Debug(dat)

	return dat["timestamp"].(string),dat["sender"].(string),dat["message"].(string)
}

//Create json before sending it to message manager
func CreateJson(timestamp string,sender string,message string) (string) {

	Debug("timestamp: "+timestamp)
	Debug("sender: "+sender)
	Debug("message: "+message)
	sMessage:= map[string]string{"timestamp": timestamp, "sender": sender,"message":message}
	json, err := json.Marshal(sMessage)

	if err!= nil{
		Error("Error converting message into json")
	}

	Debug("createJson: "+string(json))

	return string(json)
}
