package main

import (
	"flag"
	"fmt"
	"time"
)

const (
	MODE_PRODUCE = "produce"
	MODE_CONSUME = "consume"
)

func main() {

	mode := flag.String("mode", "produce", "Produce will run binlogger and dispatchs events; consume will run consumer")
	exchange := flag.String("exchange", "binlog", "Exchange to pub/sub")
	flag.Parse()

	switch *mode {
	case MODE_PRODUCE: //todo handle forever running

		go run(*exchange)

		time.Sleep(2 * time.Minute)
		fmt.Print("Shutting down...")

	case MODE_CONSUME:
		consume(*exchange)
	default:
		panic(fmt.Sprintf("No mode for %s", *mode))
	}

}
