package main

import (
	"log"
	"yandexlyceum/internal/application"
)

func main() {
	agent := application.NewAgent()
	log.Println("Starting Agent...")
	agent.Run()
}
