package main

import (
	"encoding/json"
	"fmt"
	"ka-ping/internal/config"
	"ka-ping/internal/device"
	"ka-ping/internal/netinfo"
	"ka-ping/internal/sender"
	"ka-ping/internal/ui"
	"log"
	"net/http"
	"time"
)

type ClientStatus struct {
	LastPing     time.Time          `json:"last_ping"`
	LastData     *device.DeviceInfo `json:"last_data"`
	ServerStatus string             `json:"server_status"`
	PingCount    int                `json:"ping_count"`
	Errors       []string           `json:"errors"`
}

var status ClientStatus

func main() {
	cfg := config.Load()
	dev := device.NewDevice(cfg)
	net := netinfo.NewNetInfo()
	s := sender.NewSender(cfg)

	// Start HTTP server for client interface
	go startClientInterface()

	// Main ping loop
	for {
		info := dev.Collect()
		info.PublicIP, info.Geo, info.Latitude, info.Longitude = net.GetPublicIPAndGeo()

		log.Printf("Collected device info: Hostname=%s, OS=%s, IP=%s, Lat=%s, Lon=%s",
			info.Hostname, info.OS, info.PublicIP, info.Latitude, info.Longitude)

		resp, err := s.Send(info)
		if err != nil {
			log.Printf("Error sending data: %v", err)
			status.Errors = append(status.Errors, fmt.Sprintf("%s: %v", time.Now().Format("15:04:05"), err))
			status.ServerStatus = "Error"
		} else {
			log.Printf("Server response: %s", resp)
			status.ServerStatus = "Connected"
		}

		status.LastPing = time.Now()
		status.LastData = info
		status.PingCount++

		time.Sleep(5 * time.Minute)
	}
}

func startClientInterface() {
	http.HandleFunc("/", serveClientInterface)
	http.HandleFunc("/api/status", serveStatus)

	log.Println("Client interface starting on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func serveClientInterface(w http.ResponseWriter, r *http.Request) {
	if err := ui.RenderIndex(w); err != nil {
		log.Printf("client UI render error: %v", err)
		http.Error(w, "failed to render client interface", http.StatusInternalServerError)
	}
}

func serveStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
