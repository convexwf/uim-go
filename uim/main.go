// Copyright 2024 convexwf
//
// Project: uim-go
// File: main.go
// Email: convexwf@gmail.com
// Created: 2024-10-09
// Last modified: 2024-01-22
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Description: This is a simple echo server using gin and gorilla/websocket.

package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:18080", "http service address")

var upgrader = websocket.Upgrader{} // use default options

// echo handles WebSocket requests by echoing the request message back to the client.
func echo(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Fatalf("upgrade error: %v", err)
	}
	defer conn.Close()
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("read error: %v", err)
		}
		log.Printf("recv: %s", message)
		err = conn.WriteMessage(messageType, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	router := gin.Default()
	router.GET("/echo", echo)

	if err := http.ListenAndServe(*addr, router); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
