package netinfo

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type NetInfo struct{}

func NewNetInfo() *NetInfo {
	return &NetInfo{}
}

func (n *NetInfo) GetPublicIPAndGeo() (string, map[string]string, string, string) {
	resp, err := http.Get("http://ip-api.com/json/")
	if err != nil {
		return "", nil, "", ""
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	ip, _ := result["query"].(string)
	geo := map[string]string{
		"country": fmt.Sprintf("%v", result["country"]),
		"region":  fmt.Sprintf("%v", result["regionName"]),
		"city":    fmt.Sprintf("%v", result["city"]),
		"lat":     fmt.Sprintf("%v", result["lat"]),
		"lon":     fmt.Sprintf("%v", result["lon"]),
	}

	lat := fmt.Sprintf("%v", result["lat"])
	lon := fmt.Sprintf("%v", result["lon"])

	return ip, geo, lat, lon
}
