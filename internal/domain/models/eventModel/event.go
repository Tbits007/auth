package eventModel

import "github.com/google/uuid"


type EventStatus string

const (
    PENDING   EventStatus = "pending"
    PROCESSED EventStatus = "processed"
	FAILED 	  EventStatus = "failed"
)

type Event struct {
	ID 			 uuid.UUID 
	EventType    string
	Payload 	 []byte
	Status       EventStatus 
}