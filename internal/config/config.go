package config

type Config struct {
	Endpoint string
	UUID     string
}

func Load() *Config {
	return &Config{
		Endpoint: "http://localhost:8080/api/device-ping",
		UUID:     "", // Will be set by device.NewDevice
	}
}
