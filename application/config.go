package application

import (
	"os"
	"strconv"
)

type config struct {
	redis_address string
	server_port   uint16
}

func loadconfig() config {
	cfg := config{
		redis_address: "localhost:6379",
		server_port:   3000,
	}
	if redisAddr, exists := os.LookupEnv("REDIS_ADDR"); exists {
		cfg.redis_address = redisAddr
	}
	if servePort, exists := os.LookupEnv("SERVER"); exists {
		if port, err := strconv.ParseUint(servePort, 10, 16); err == nil {
			cfg.server_port = uint16(port)
		}
	}
	return cfg

}
