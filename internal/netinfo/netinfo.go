package netinfo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type NetInfo struct {
	client *http.Client
}

func NewNetInfo() *NetInfo {
	return &NetInfo{
		client: &http.Client{
			Timeout: 8 * time.Second,
		},
	}
}

func (n *NetInfo) GetPublicIPAndGeo() (string, map[string]string, string, string) {
	setGeo := func(m *map[string]string, key, value string) {
		if value == "" {
			return
		}
		if *m == nil {
			*m = make(map[string]string)
		}
		(*m)[key] = value
	}

	joinNonEmpty := func(parts ...string) string {
		var cleaned []string
		for _, p := range parts {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				cleaned = append(cleaned, trimmed)
			}
		}
		return strings.Join(cleaned, ", ")
	}

	var (
		ip       string
		lat      string
		lon      string
		geo      map[string]string
		ipAPIRes *ipAPIResponse
	)

	// First try to discover the public IP using a dedicated endpoint.
	if lookupIP, err := n.lookupPublicIP(); err == nil {
		ip = lookupIP
	}

	// Keep an ip-api.com result handy as a fallback for richer geo details.
	if res, err := n.lookupIPAPI(); err == nil {
		ipAPIRes = res
		if ip == "" {
			ip = res.Query
		}
	}

	// Primary geo lookup through ipwho.is to obtain accurate coordinates and metadata.
	if ip != "" {
		if ipwhoRes, err := n.lookupIPWho(ip); err == nil && ipwhoRes != nil && ipwhoRes.Success {
			if ipwhoRes.IP != "" {
				ip = ipwhoRes.IP
			}
			if ipwhoRes.Latitude != 0 && ipwhoRes.Longitude != 0 {
				lat = fmt.Sprintf("%.6f", ipwhoRes.Latitude)
				lon = fmt.Sprintf("%.6f", ipwhoRes.Longitude)
			}

			setGeo(&geo, "country", ipwhoRes.Country)
			setGeo(&geo, "country_code", ipwhoRes.CountryCode)
			setGeo(&geo, "region", ipwhoRes.Region)
			setGeo(&geo, "city", ipwhoRes.City)
			setGeo(&geo, "postal", ipwhoRes.Postal)
			setGeo(&geo, "continent", ipwhoRes.Continent)
			setGeo(&geo, "continent_code", ipwhoRes.ContinentCode)
			setGeo(&geo, "timezone", ipwhoRes.Timezone.ID)
			setGeo(&geo, "timezone_abbr", ipwhoRes.Timezone.Abbreviation)
			setGeo(&geo, "timezone_offset", ipwhoRes.Timezone.UTC)
			setGeo(&geo, "ip_type", ipwhoRes.Type)

			if place := joinNonEmpty(ipwhoRes.City, ipwhoRes.Region, ipwhoRes.Country); place != "" {
				setGeo(&geo, "place_name", place)
			}
			if ipwhoRes.Connection.ASN != 0 {
				setGeo(&geo, "asn", fmt.Sprintf("%d", ipwhoRes.Connection.ASN))
			}
			setGeo(&geo, "organization", ipwhoRes.Connection.Organization)
			setGeo(&geo, "isp", ipwhoRes.Connection.ISP)
			setGeo(&geo, "network_domain", ipwhoRes.Connection.Domain)
			setGeo(&geo, "lookup_source", "ipwho.is")
			setGeo(&geo, "lookup_timestamp_utc", time.Now().UTC().Format(time.RFC3339))
		}
	}

	// Fallback to ip-api.com data if primary lookup failed or missed fields.
	if (lat == "" || lon == "") || geo == nil {
		if ipAPIRes != nil && ipAPIRes.Status == "success" {
			if lat == "" && lon == "" && ipAPIRes.Lat != 0 && ipAPIRes.Lon != 0 {
				lat = fmt.Sprintf("%.6f", ipAPIRes.Lat)
				lon = fmt.Sprintf("%.6f", ipAPIRes.Lon)
			}

			if geo == nil {
				geo = make(map[string]string)
			}
			setGeo(&geo, "country", ipAPIRes.Country)
			setGeo(&geo, "country_code", ipAPIRes.CountryCode)
			setGeo(&geo, "region", ipAPIRes.RegionName)
			setGeo(&geo, "region_code", ipAPIRes.Region)
			setGeo(&geo, "city", ipAPIRes.City)
			setGeo(&geo, "postal", ipAPIRes.Zip)
			setGeo(&geo, "timezone", ipAPIRes.Timezone)
			setGeo(&geo, "organization", ipAPIRes.Org)
			setGeo(&geo, "isp", ipAPIRes.ISP)
			setGeo(&geo, "asn", ipAPIRes.AS)
			if _, exists := geo["place_name"]; !exists {
				if place := joinNonEmpty(ipAPIRes.City, ipAPIRes.RegionName, ipAPIRes.Country); place != "" {
					setGeo(&geo, "place_name", place)
				}
			}
			setGeo(&geo, "lookup_source", "ip-api.com")
			if _, exists := geo["lookup_timestamp_utc"]; !exists {
				setGeo(&geo, "lookup_timestamp_utc", time.Now().UTC().Format(time.RFC3339))
			}
		}
	}

	if lat != "" && lon != "" {
		setGeo(&geo, "google_maps_url", fmt.Sprintf("https://maps.google.com/?q=%s,%s", lat, lon))
	}

	return ip, geo, lat, lon
}

func (n *NetInfo) lookupPublicIP() (string, error) {
	var resp ipifyResponse
	if err := n.fetchJSON("https://api.ipify.org?format=json", &resp); err != nil {
		return "", err
	}
	return resp.IP, nil
}

func (n *NetInfo) lookupIPWho(ip string) (*ipWhoResponse, error) {
	var resp ipWhoResponse
	if err := n.fetchJSON(fmt.Sprintf("https://ipwho.is/%s", ip), &resp); err != nil {
		return nil, err
	}
	if !resp.Success {
		return &resp, fmt.Errorf("ipwho.is lookup failed: %s", resp.Message)
	}
	return &resp, nil
}

func (n *NetInfo) lookupIPAPI() (*ipAPIResponse, error) {
	var resp ipAPIResponse
	if err := n.fetchJSON("http://ip-api.com/json/", &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (n *NetInfo) fetchJSON(url string, target interface{}) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := n.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("request failed: %s", resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

type ipifyResponse struct {
	IP string `json:"ip"`
}

type ipWhoResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	IP      string `json:"ip"`
	Type    string `json:"type"`

	Continent     string  `json:"continent"`
	ContinentCode string  `json:"continent_code"`
	Country       string  `json:"country"`
	CountryCode   string  `json:"country_code"`
	Region        string  `json:"region"`
	City          string  `json:"city"`
	Postal        string  `json:"postal"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`

	Timezone struct {
		ID           string `json:"id"`
		Abbreviation string `json:"abbr"`
		UTC          string `json:"utc"`
	} `json:"timezone"`

	Connection struct {
		ASN          int    `json:"asn"`
		Organization string `json:"org"`
		ISP          string `json:"isp"`
		Domain       string `json:"domain"`
	} `json:"connection"`
}

type ipAPIResponse struct {
	Status      string  `json:"status"`
	Message     string  `json:"message"`
	Query       string  `json:"query"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"region"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Zip         string  `json:"zip"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Timezone    string  `json:"timezone"`
	ISP         string  `json:"isp"`
	Org         string  `json:"org"`
	AS          string  `json:"as"`
}
