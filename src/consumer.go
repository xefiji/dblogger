package main

import (
	"encoding/json"
	"log"
	"regexp"

	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
)

var client Rmq

//consume
func consume(exchange string) {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	//connect to broker
	client = Rmq{}
	client.Connect(exchange)

	msgs := client.Receive(exchange)
	forever := make(chan bool)

	if exchangeIsInternal(exchange) {
		consumeForElastic(msgs)
	}

	if exchangeIsPublic(exchange) {
		consumeStream(msgs)
	}

	log.Printf(" [*] Waiting for logs. To exit press CTRL+C")
	<-forever

	log.Println("✅ Consumer shut down gracefully")
}

//exchangeIsInternal
func exchangeIsInternal(exchange string) bool {
	var re = regexp.MustCompile(`(?m)_internal$`)
	return re.Match([]byte(exchange))
}

//exchangeIsPublic
func exchangeIsPublic(exchange string) bool {
	var re = regexp.MustCompile(`(?m)_public$`)
	return re.Match([]byte(exchange))
}

//consumeStream only for some tests with workers that consume real time events streaming
//in real world this should be made by other clients app
func consumeStream(msgs <-chan amqp.Delivery) {
	for d := range msgs {
		log.Printf(" [x] %s", d.Body)
	}
}

//consumeForElastic gets all messages from direct exchange (in a FIFO queue)
//and must index them in elastic search
func consumeForElastic(msgs <-chan amqp.Delivery) {
	for d := range msgs {
		event := Event{}
		json.Unmarshal(d.Body, &event)
		pretty(event)
	}
}
