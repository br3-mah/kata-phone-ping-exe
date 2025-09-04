package main

import (
	"encoding/json"
	"fmt"
	"ka-ping/internal/config"
	"ka-ping/internal/device"
	"ka-ping/internal/netinfo"
	"ka-ping/internal/sender"
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
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, clientInterfaceHTML)
}

func serveStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

const clientInterfaceHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Ka-Ping Client Monitor</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: Arial, sans-serif; background-color: #f5f5f5; }
        .container { max-width: 1000px; margin: 0 auto; padding: 20px; }
        .header { text-align: center; margin-bottom: 30px; }
        .header h1 { color: #333; margin-bottom: 10px; }
        .status-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 20px; margin-bottom: 30px; }
        .status-card { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .status-card h3 { color: #333; margin-bottom: 15px; }
        .status-value { font-size: 1.2em; font-weight: bold; }
        .status-connected { color: #28a745; }
        .status-error { color: #dc3545; }
        .data-section { background: white; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); padding: 20px; margin-bottom: 20px; }
        .data-section h3 { color: #333; margin-bottom: 15px; }
        .data-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 15px; }
        .data-item { padding: 10px; background: #f8f9fa; border-radius: 4px; }
        .data-label { font-weight: bold; color: #666; font-size: 0.9em; }
        .data-value { margin-top: 5px; color: #333; }
        .refresh-btn { background: #007bff; color: white; border: none; padding: 10px 20px; border-radius: 4px; cursor: pointer; margin-bottom: 20px; }
        .refresh-btn:hover { background: #0056b3; }
        .errors-section { background: white; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); padding: 20px; }
        .error-item { padding: 8px; background: #f8d7da; color: #721c24; border-radius: 4px; margin-bottom: 5px; }
        .map-link { color: #007bff; text-decoration: none; }
        .map-link:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Ka-Ping Client Monitor</h1>
            <p>Real-time monitoring of device data being sent to server</p>
        </div>

        <button class="refresh-btn" onclick="loadStatus()">Refresh</button>

        <div class="status-grid">
            <div class="status-card">
                <h3>Server Status</h3>
                <div class="status-value" id="serverStatus">Loading...</div>
            </div>
            <div class="status-card">
                <h3>Last Ping</h3>
                <div class="status-value" id="lastPing">Loading...</div>
            </div>
            <div class="status-card">
                <h3>Ping Count</h3>
                <div class="status-value" id="pingCount">Loading...</div>
            </div>
        </div>

        <div class="data-section">
            <h3>Last Sent Device Data</h3>
            <div class="data-grid" id="deviceData">
                <div class="data-item">
                    <div class="data-label">Loading...</div>
                    <div class="data-value">Please wait...</div>
                </div>
            </div>
        </div>

        <div class="errors-section">
            <h3>Recent Errors</h3>
            <div id="errorsList">
                <div class="error-item">No errors</div>
            </div>
        </div>
    </div>

    <script>
        async function loadStatus() {
            try {
                const response = await fetch('/api/status');
                const status = await response.json();

                // Update status cards
                document.getElementById('serverStatus').textContent = status.server_status;
                document.getElementById('serverStatus').className = 'status-value ' +
                    (status.server_status === 'Connected' ? 'status-connected' : 'status-error');

                document.getElementById('lastPing').textContent = status.last_ping ?
                    new Date(status.last_ping).toLocaleString() : 'Never';

                document.getElementById('pingCount').textContent = status.ping_count || 0;

                // Update device data
                if (status.last_data) {
                    const data = status.last_data;
                    const location = data.geo ?
                        [data.geo.city, data.geo.region, data.geo.country].filter(Boolean).join(', ') :
                        'Unknown';
                    
                    const mapLink = (data.latitude && data.longitude) ?
                        '<a href="https://maps.google.com/?q=' + data.latitude + ',' + data.longitude + '" target="_blank" class="map-link">View on Google Maps</a>' :
                        'Coordinates not available';

                    document.getElementById('deviceData').innerHTML = 
                        '<div class="data-item">' +
                            '<div class="data-label">Hostname</div>' +
                            '<div class="data-value">' + data.hostname + '</div>' +
                        '</div>' +
                        '<div class="data-item">' +
                            '<div class="data-label">UUID</div>' +
                            '<div class="data-value">' + data.uuid.substring(0, 8) + '...</div>' +
                        '</div>' +
                        '<div class="data-item">' +
                            '<div class="data-label">Operating System</div>' +
                            '<div class="data-value">' + data.os + '</div>' +
                        '</div>' +
                        '<div class="data-item">' +
                            '<div class="data-label">MAC Address</div>' +
                            '<div class="data-value">' + data.mac + '</div>' +
                        '</div>' +
                        '<div class="data-item">' +
                            '<div class="data-label">Public IP</div>' +
                            '<div class="data-value">' + data.public_ip + '</div>' +
                        '</div>' +
                        '<div class="data-item">' +
                            '<div class="data-label">Location</div>' +
                            '<div class="data-value">' + location + '</div>' +
                        '</div>' +
                        '<div class="data-item">' +
                            '<div class="data-label">Coordinates</div>' +
                            '<div class="data-value">' + (data.latitude || 'N/A') + ', ' + (data.longitude || 'N/A') + '</div>' +
                        '</div>' +
                        '<div class="data-item">' +
                            '<div class="data-label">Google Maps</div>' +
                            '<div class="data-value">' + mapLink + '</div>' +
                        '</div>';
                }
                // Update errors
                const errorsList = document.getElementById('errorsList');
                if (status.errors && status.errors.length > 0) {
                    errorsList.innerHTML = status.errors.map(function(error) {
                        return '<div class="error-item">' + error + '</div>';
                    }).join('');
                } else {
                    errorsList.innerHTML = '<div class="error-item">No errors</div>';
                }
            } catch (error) {
                console.error('Error loading status:', error);
            }
        }

        // Load status on page load
        loadStatus();

        // Auto-refresh every 10 seconds
        setInterval(loadStatus, 10000);
    </script>
</body>
</html>
`
