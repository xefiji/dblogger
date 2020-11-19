package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
)

var client Rmq

//consume
func consume(exchange string) {

	log.Printf("[consumer] consumer started from exchange %s ✅\n", exchange)

	err := godotenv.Load()
	if err != nil {
		log.Fatal("[consumer] Error loading .env file ❌")
	}

	//connect to broker
	client = Rmq{}
	client.Connect(exchange)

	msgs := client.Receive(exchange)
	quit := make(chan os.Signal)

	if exchangeIsInternal(exchange) {

		es := GetESInstance()
		if false == es.IndexExists(DEFAULT_INDEX) {
			es.Map(DEFAULT_INDEX, "mapping.json")
		}

		go consumeForElastic(msgs)
	}

	if exchangeIsPublic(exchange) {
		go consumeStream(msgs)
	}

	log.Printf(" [*] Waiting for logs. To exit press CTRL+C")
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[consumer] shut down gracefully ✅")
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

		s, err := json.Marshal(event)
		if err != nil {
			log.Fatalf("[consumer] could not marshal event: %s ❌", err.Error())
			continue
		}

		GetESInstance().IndexOne(DEFAULT_INDEX, event.Id, strings.NewReader(string(s)))
	}
}
