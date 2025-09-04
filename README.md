# ka-ping

A small Windows background service that pings device info every 5 minutes.

## Build

1. Install Go (https://golang.org/dl/)
2. Open terminal in project root.
3. Run:
   ```
   go mod tidy
   go build -ldflags "-H windowsgui" -o ka-ping.exe ./cmd/ka-ping
   ```

## Install

- Place `ka-ping.exe` in your Startup folder:
  ```
  %APPDATA%\Microsoft\Windows\Start Menu\Programs\Startup
  ```
- It will run silently in the background on startup.

## Features

- Collects hostname, OS, MAC, public IP, geolocation.
- Sends JSON to https://example.com/api/device-ping every 5 minutes.
- Uses a persistent UUID for device identification. 

Overview of ka-ping.exe

This application is a small Windows service written in Go that performs device information pinging. Here's a concise breakdown:

#### Purpose
•  Collector: Gathers device information like hostname, OS, MAC address, public IP, and geolocation.
•  Sender: Sends this data every five minutes to a specified endpoint: https://example.com/api/device-ping.

#### Key Components
•  Configuration (config.go): Loads endpoint URL and UUID configuration.
•  Device Information (device.go): Collects device details including UUID, hostname, OS, and MAC address.
•  UUID Generation (uuid.go): Uses Google's UUID library to create or retrieve a persistent UUID stored in the APPDATA directory.
•  MAC Address Retrieval (mac.go): Extracts MAC address from network interfaces.
•  Main Logic (main.go): Orchestrates the process of gathering information and sending it to the server every five minutes.

How It Works
1. Startup: Initialize configuration and create a new device instance.
2. Device Info Collection: Gather host details and MAC address.
3. Public IP and Geo: Fetch public IP and geolocation info (specific implementation details likely in netinfo and ping packages, which may need further exploration).
4. Send Info: Transmit the collected data to the remote server.
5. Repeat Every 5 Minutes: Continually collect and send device information.

Great! Your server is running and the database is ready. Now, let's ensure you can **see logs from the ka-ping client** to confirm it's collecting and sending data.

### 1. Add Logging to ka-ping

Currently, the ka-ping client runs silently (no console window, no logs).  
To see logs (for debugging or monitoring), you have two options:

#### **Option A: Run in Console Mode for Debugging**

- Temporarily build and run ka-ping **without** the `-ldflags "-H windowsgui"` flag:
  ```sh
  go build -o ka-ping-debug.exe ./cmd/ka-ping
  .\ka-ping-debug.exe
  ```
- This will show logs in the terminal window.

#### **Option B: Add Logging to the Code**

Update your code to print logs for each step (collecting info, sending data, response from server, errors, etc.).  
Here’s how you can add logging to your main loop and sender:

---

**Edit `cmd/ka-ping/main.go`:**

```go
package main

import (
	"fmt"
	"time"
	"ka-ping/internal/config"
	"ka-ping/internal/device"
	"ka-ping/internal/netinfo"
	"ka-ping/internal/sender"
)

func main() {
	cfg := config.Load()
	dev := device.NewDevice(cfg)
	net := netinfo.NewNetInfo()
	s := sender.NewSender(cfg)

	for {
		info := dev.Collect()
		info.PublicIP, info.Geo = net.GetPublicIPAndGeo()
		fmt.Printf("Collected info: %+v\n", info)
		resp, err := s.Send(info)
		if err != nil {
			fmt.Printf("Error sending data: %v\n", err)
		} else {
			fmt.Printf("Server response: %s\n", resp)
		}
		time.Sleep(5 * time.Minute)
	}
}
```

**Edit `internal/sender/sender.go`:**

```go
func (s *Sender) Send(info *device.DeviceInfo) (string, error) {
	data, _ := json.Marshal(info)
	resp, err := http.Post(s.cfg.Endpoint, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return string(body), nil
}
```

---

### 2. Build and Run in Console

- Build and run as above (without `-H windowsgui`), so you can see the logs:
  ```sh
  go build -o ka-ping-debug.exe ./cmd/ka-ping
  .\ka-ping-debug.exe
  ```

### 3. Check the Server

- You should see log output in the terminal for each POST.
- On the server, you should see new records in the `devices` table and logs in the server terminal.

---

### 4. Once Confirmed

- When you’re satisfied, rebuild with `-ldflags "-H windowsgui"` for silent background operation.

---

**Let me know if you want me to generate the exact code edits for logging, or if you want to proceed with the above steps!**