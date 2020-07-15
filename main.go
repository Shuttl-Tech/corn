package main

import (
	"github.com/Shuttl-Tech/corn/cmd"
	"log"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		log.Fatalf("command failed with error %s", err)
	}
}
