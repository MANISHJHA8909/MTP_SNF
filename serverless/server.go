package main

import (
	"encoding/json"
	"net/http"

	"github.com/redis/go-redis/v9"
)

var rdb = redis.NewClient(&redis.Options{
	Addr: "redis:6379",
})

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Type string `json:"type"`
			MAC  string `json:"mac"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		switch req.Type {
		case "DISCOVER":
			ip := allocateIP(req.MAC)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"offer":      ip,
				"lease_time": 3600,
			})
		case "REQUEST":
			json.NewEncoder(w).Encode(map[string]bool{"ack": true})
		}
	})
	http.ListenAndServe(":8080", nil)
}

func allocateIP(mac string) string {
	// Implement IP allocation logic with Redis
	return "192.168.1.100"
}
