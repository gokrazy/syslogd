// Binary gsl is a front-end for accessing the gokrazy syslog.
package main

import (
	"flag"
	"log"
)

func gsl() error {
	log.Printf("TODO: implement gsl")
	return nil
}

func main() {
	flag.Parse()
	if err := gsl(); err != nil {
		log.Fatal(err)
	}
}
