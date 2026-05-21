package main

import (
	"log"

	"gitlab.com/marsskom/burro/internal/cert"
)

func main() {
	err := cert.GenerateCA(
		"./certs/ca.pem",
		"./certs/ca.key",
	)
	if err != nil {
		log.Fatal(err)
	}
}
