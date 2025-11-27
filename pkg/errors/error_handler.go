package errors

import (
	"log"
	"os"
)

type AppError struct {
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
	Err        error  `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

func New(message string, statusCode int, err error) *AppError {
	return &AppError{
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
	}
}

func (e *AppError) Log() {
	errorLogger := log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	if e.Err != nil {
		errorLogger.Println(e.Message, e.Err)
	} else {
		errorLogger.Println(e.Message)
	}
}
