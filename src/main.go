package main

import (
	"flag"
	"fmt"
)

const (
	MODE_PRODUCE = "produce"
	MODE_CONSUME = "consume"
)

//todo internal exchange should be a different user
func main() {

	mode := flag.String("mode", "produce", "Produce will run binlogger and dispatchs events; consume will run consumer")
	exchange := flag.String("exchange", "binlog", "Exchange to pub/sub")
	flag.Parse()

	switch *mode {

	case MODE_PRODUCE:
		produce(*exchange)

	case MODE_CONSUME:
		consume(*exchange)

	default:
		panic(fmt.Sprintf("No mode for %s", *mode))
	}
}
