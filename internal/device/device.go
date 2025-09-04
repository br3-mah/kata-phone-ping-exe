package device

import (
	"ka-ping/internal/config"
	"os"
	"runtime"
)

type DeviceInfo struct {
	UUID      string            `json:"uuid"`
	Hostname  string            `json:"hostname"`
	OS        string            `json:"os"`
	MAC       string            `json:"mac"`
	PublicIP  string            `json:"public_ip"`
	Geo       map[string]string `json:"geo"`
	Latitude  string            `json:"latitude"`
	Longitude string            `json:"longitude"`
}

type Device struct {
	uuid string
}

func NewDevice(cfg *config.Config) *Device {
	uuid := GetOrCreateUUID()
	cfg.UUID = uuid
	return &Device{uuid: uuid}
}

func (d *Device) Collect() *DeviceInfo {
	hostname, _ := os.Hostname()
	return &DeviceInfo{
		UUID:     d.uuid,
		Hostname: hostname,
		OS:       runtime.GOOS + " " + runtime.GOARCH,
		MAC:      getMAC(),
	}
}
