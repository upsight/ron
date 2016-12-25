package main

import (
	"log"
	"os"

	"github.com/upsight/ron"
	"github.com/upsight/ron/color"
)

func main() {
	status, err := ron.Run(os.Stdout, os.Stderr, os.Args[1:])
	if err != nil {
		hostname, _ := os.Hostname()
		log.Println(hostname, color.Red(err.Error()))
	}
	os.Exit(status)
}
