package main

type Event struct {
	Readable string
	Action   string
	Schema   string
	Table    string
	Payload  map[string]interface{}
	Origin   map[string]interface{}
	Header   EventHeader
}

type EventHeader struct {
	Timestamp uint32
	EventType string
	ServerID  uint32
	EventSize uint32
}
