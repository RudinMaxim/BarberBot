package config

import (
	"log"
	"os"
	"sync"
)

var (
	actionLogger *log.Logger
	once         sync.Once
)

func init() {
	once.Do(func() {
		actionFile, err := os.OpenFile("actions.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		actionLogger = log.New(actionFile, "ACTION: ", log.Ldate|log.Ltime|log.Lshortfile)
	})
}

func LogAction(message string) {
	actionLogger.Println(message)
}
