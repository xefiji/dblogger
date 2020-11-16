package main

import (
	"log"

	"github.com/joho/godotenv"
)

func consume(exchange string) {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	//connect to broker
	client := Rmq{}
	client.Connect(exchange)

	msgs := client.Receive(exchange)

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf(" [x] %s", d.Body)
		}
	}()

	log.Printf(" [*] Waiting for logs. To exit press CTRL+C")
	<-forever
}
