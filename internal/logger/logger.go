package logger

import "log"

type Logger struct{}

func New() *Logger {
	return &Logger{}
}

func (l *Logger) Error(message string, err error) {
	if err == nil {
		log.Println(message)
		return
	}
	log.Printf("%s: %v\n", message, err)
}
