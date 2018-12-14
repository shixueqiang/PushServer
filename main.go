package main

/*
#cgo CFLAGS: -I${SRCDIR}/include
#cgo LDFLAGS: -L${SRCDIR}/lib -lpushmqtt -lmsgpackc
#include <stdlib.h>
extern int mqtt_main(int argc, char *argv[]);
extern int mqtt_publish(const char *topic, const char *payload, int qos);
*/
import "C"

import (
	"log"
	"net/http"
	"time"
	"unsafe"

	"github.com/gin-gonic/gin"
)

//export CGOMessageCallback
func CGOMessageCallback(msgType uint8, msgTimestamp uint32, clientId *C.char, topic *C.char, msgId *C.char, payload *C.char) {
	log.Printf("MessageCallback msgType:%v msgTimestamp:%v clientId:%v topic:%v msgId:%v payload:%v\n", msgType, msgTimestamp, C.GoString(clientId), C.GoString(topic), C.GoString(msgId), C.GoString(payload))
}

func Intercepter() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()

		// before request
		log.Printf("befor request %s", c.Request.RequestURI)

		c.Next()

		// after request
		latency := time.Since(t)
		log.Print(latency)

		// access the status we are sending
		status := c.Writer.Status()
		log.Printf("after request %s status is %d", c.Request.RequestURI, status)
	}
}

func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()
	r.Use(Intercepter())

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
		topic := C.CString("test/topic")
		payload := C.CString("from golang")
		defer C.free(unsafe.Pointer(topic))
		defer C.free(unsafe.Pointer(payload))
		C.mqtt_publish(topic, payload, C.int(1))
	})

	return r
}

func main() {
	//开启mqtt订阅
	args := []string{"mosquitto_sub", "-h", "127.0.0.1", "-t", "topic/push", "-i", "client_push", "-q", "1", "-c", "-d"}
	argc := C.int(len(args))
	argv := make([]*C.char, argc)
	for i, s := range args {
		argv[i] = C.CString(s)
		defer C.free(unsafe.Pointer(argv[i]))
	}
	go C.mqtt_main(argc, &argv[0])
	r := setupRouter()
	// Listen and Server in 0.0.0.0:8080
	r.Run(":8080")
}
