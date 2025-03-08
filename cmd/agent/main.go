package main

import (
	"log"
	"yandexlyceum/yandex_lyceum_go/internal/application"
)

func main() {
	agent := application.NewAgent()
	log.Println("Starting Agent...")
	agent.Run()
}
