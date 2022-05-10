package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
)

const Host = "192.168.1.242"
const Port = "38899"

type Params struct {
	State     bool    `json:"state"`
	WarmWhite float64 `json:"w,omitempty"`       // value > 0 < 256
	ColdWhite float64 `json:"c,omitempty"`       // value > 0 < 256
	Speed     int64   `json:"speed,omitempty"`   // value > 0 < 101
	SceneId   int64   `json:"sceneId,omitempty"` // SceneModel
	R         float64 `json:"r,omitempty"`
	G         float64 `json:"g,omitempty"`
	B         float64 `json:"b,omitempty"`
	Dimming   int64   `json:"dimming,omitempty"`
	Temp      float64 `json:"temp,omitempty"`
}

type Payload struct {
	Method string `json:"method,omitempty"`
	Params Params `json:"params,omitempty"`
}

type Response struct {
	Success     bool    `json:"success,omitempty"`
	Mac         string  `json:"mac,omitempty"`
	Rssi        int64   `json:"rssi,omitempty"`
	Src         string  `json:"src,omitempty"`
	State       bool    `json:"state,omitempty"`
	SceneId     int64   `json:"sceneId,omitempty"`
	Speed       int64   `json:"speed,omitempty"`
	Temp        int64   `json:"temp,omitempty"`
	Dimming     int64   `json:"dimming,omitempty"`
	HomeId      int64   `json:"homeId,omitempty"`
	RoomId      int64   `json:"roomId,omitempty"`
	HomeLock    bool    `json:"homeLock,omitempty"`
	PairingLock bool    `json:"pairingLock,omitempty"`
	TypeId      int64   `json:"typeId,omitempty"`
	ModuleName  string  `json:"module_name,omitempty"`
	FwVersion   string  `json:"fwVersion,omitempty"`
	GroupId     int64   `json:"groupId,omitempty"`
	DrvConf     []int64 `json:"drvConf,omitempty"`
}

func sendMessage(host string, message *Payload) *Response {
	remoteAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(`%s:%s`, host, Port))
	if err != nil {
		log.Fatal("ResolveUDPAddr:", err)
	}
	conn, err := net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		log.Fatal("DialUDP", err)
	}
	defer conn.Close()
	// marshall payload to json string
	payload, err := json.Marshal(message)
	if err != nil {
		log.Fatal("Marshal", err)
	}
	payloadString := string(payload)
	fmt.Println(fmt.Sprintf(`Payload string: %s`, payloadString))
	// send payload to bulb
	if _, err = conn.Write(payload); err != nil {
		log.Fatal("Write", err)
	}
	// read response from bulb
	var responsePayload = new(Response)
	var response = make([]byte, 4096)
	if _, err = bufio.NewReader(conn).Read(response); err != nil {
		log.Fatal("Read:", err)
	}
	result := []byte(strings.Trim(string(response), "\x00'"))
	// convert string result to struct again
	if err = json.Unmarshal(result, responsePayload); err != nil {
		log.Fatal("Unmarshal", err)
	}
	fmt.Println(responsePayload)
	return responsePayload
}

func main() {
	logCmd := exec.Command("log", "stream", "--predicate", `eventMessage contains "Post event kCameraStream"`)
	logOut, err := logCmd.StdoutPipe()
	if err != nil {
		log.Fatal("StdoutPipe:", err)
	}
	reader := bufio.NewReader(logOut)
	if err = logCmd.Start(); err != nil {
		log.Fatal("Buffer Error:", err)
	}
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("Read Error:", err)
			return
		}
		fmt.Println(line)
		if strings.Contains(line, "kCameraStreamStart") {
			fmt.Println("LIGHT ON")
			sendMessage(Host, &Payload{
				Method: "setPilot",
				Params: Params{
					State: true,
					Speed: 50,
					R:     255,
					G:     0,
					B:     0,
				},
			})
		}
		if strings.Contains(line, "kCameraStreamStop") {
			fmt.Println("LIGHT OFF")
			sendMessage(Host, &Payload{
				Method: "setPilot",
				Params: Params{
					State: false,
					Speed: 50,
				},
			})
		}
	}
}
