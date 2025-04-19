// utils/logger.go
package utils

import (
	"log"
	"os"
)

var Logger *log.Logger

func InitLogger() {
	Logger = log.New(os.Stdout, "[JSCol] ", log.LstdFlags|log.Lshortfile)
}
