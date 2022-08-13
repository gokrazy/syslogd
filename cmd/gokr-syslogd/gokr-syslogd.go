package main

import (
	"flag"
	"log"
)

func gokrsyslogd() error {
	log.Printf("hoi")
	return nil
}

func main() {
	flag.Parse()
	if err := gokrsyslogd(); err != nil {
		log.Fatal(err)
	}
}
