package models

import "github.com/google/uuid"


type Event struct {
	ID 			 uuid.UUID 
	AggregateID  uuid.UUID 
	EventType    string
	Payload 	 []byte
	Status       string 
}