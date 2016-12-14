package main

import (
	"log"
	"os"

	"github.com/upsight/ron"
	"github.com/upsight/ron/color"
)

func main() {
	c := ron.NewDefaultCommander(os.Stdout, os.Stderr)
	status, err := ron.Run(c, os.Args[1:])
	if err != nil {
		hostname, _ := os.Hostname()
		log.Println(hostname, color.Red(err.Error()))
	}
	os.Exit(status)
}
