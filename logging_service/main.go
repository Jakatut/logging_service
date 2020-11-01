package main

import (
	log "logging_service/core"
	"logging_service/messages"
	"logging_service/routes"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var debugWaitGroup sync.WaitGroup
var warningWaitGroup sync.WaitGroup
var infoWaitGroup sync.WaitGroup
var errorWaitGroup sync.WaitGroup
var fatalWaitGroup sync.WaitGroup

func init() {
	debugWaitGroup = sync.WaitGroup{}
	warningWaitGroup = sync.WaitGroup{}
	infoWaitGroup = sync.WaitGroup{}
	errorWaitGroup = sync.WaitGroup{}
	fatalWaitGroup = sync.WaitGroup{}
}

func main() {

	logMessage := messages.Log{Message: "Hello", OriginLocation: "test", Type: 1, Severity: 1, MessageNumber: 0, CreatedDate: time.Now()}
	log.WriteLog(&logMessage)

	router := gin.New()
	routes.Setup(router, &debugWaitGroup, &warningWaitGroup, &infoWaitGroup, &errorWaitGroup, &fatalWaitGroup)
}
