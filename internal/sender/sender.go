package sender

import (
	"bytes"
	"encoding/json"
	"io"
	"ka-ping/internal/config"
	"ka-ping/internal/device"
	"net/http"
)

type Sender struct {
	cfg *config.Config
}

func NewSender(cfg *config.Config) *Sender {
	return &Sender{cfg: cfg}
}

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
