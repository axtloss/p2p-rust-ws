package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"net/http"
)

type msgInfo struct {
	Type string
	Name string
	IP   string
}

var upgrader = websocket.Upgrader{}
var info msgInfo

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func main() {

	info = msgInfo{Type: "give", Name: "", IP: GetOutboundIP().String()}

	http.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		// Upgrade upgrades the HTTP server connection to the WebSocket protocol.
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade failed: ", err)
			return
		}
		defer conn.Close()

		// Continuosly read and write message
		for {
			mt, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("read failed:", err)
				break
			}

			output := ""
			output += string(message)
			message = []byte(output)
			err = conn.WriteMessage(mt, message)
			if err != nil {
				log.Println("write failed:", err)
				break
			}
		}
	})

	http.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade failed: ", err)
			return
		}
		defer conn.Close()

		for {
			mt, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("read failed:", err)
				break
			}
			var cmd msgInfo
			var output []byte
			fmt.Println(string(message))
			err = json.Unmarshal(message, &cmd)
			if err != nil {
				log.Println("failed to parse command:", err)
				continue
			}
			if cmd.Type == "get" {
				output, err = json.Marshal(&info)
				if err != nil {
					return
				}
				fmt.Println("get")
				fmt.Println(string(output))
			} else if cmd.Type == "set" {
				info.Name = cmd.Name
				fmt.Println(info.Name)
				fmt.Println(info.IP)
			} else {
				continue
			}
			err = conn.WriteMessage(mt, output)
			if err != nil {
				log.Println("write failed:", err)
				break
			}
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "websockets.html")
	})

	http.ListenAndServe(":8080", nil)
}
