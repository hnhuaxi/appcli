package main

import (
	"log"
	"os"

	"github.com/hnhuaxi/appcli"
)

var app = appcli.App{}

func main() {
	printError(app.Execute(os.Args))
}

func printError(err error) {
	if err != nil {
		log.Printf("err: %s", err)
	}
}
