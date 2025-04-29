package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/redis/go-redis/v9"
)

var (
	rdb *redis.Client
	ctx = context.Background()
)

func init() {
	// Initialize Redis client
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis-master.redis.svc.cluster.local:6379" // Fallback
	}
	redisPassword := os.Getenv("REDIS_PASSWORD")

	rdb = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0,
	})

	// Test Redis connection
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		log.Fatalf("Failed to connect to Redis at %s: %v", redisAddr, err)
	}
	log.Println("Connected to Redis successfully")

	// Initialize IP pool if not exists
	if exists, _ := rdb.Exists(ctx, "dhcp:ip_pool_initialized").Result(); exists == 0 {
		initIPPool()
	}
}

func initIPPool() {
	startIP := os.Getenv("IP_POOL_START")
	endIP := os.Getenv("IP_POOL_END")

	if startIP == "" || endIP == "" {
		startIP = "192.168.1.100"
		endIP = "192.168.1.200"
	}

	start := ipToInt(net.ParseIP(startIP))
	end := ipToInt(net.ParseIP(endIP))

	// Store available IPs in a Redis set
	for i := start; i <= end; i++ {
		ip := intToIP(i)
		rdb.SAdd(ctx, "dhcp:available_ips", ip.String())
	}

	rdb.Set(ctx, "dhcp:ip_pool_initialized", "true", 0)
	log.Printf("Initialized IP pool from %s to %s", startIP, endIP)
}

func ipToInt(ip net.IP) int {
	ip = ip.To4()
	return int(ip[0])<<24 | int(ip[1])<<16 | int(ip[2])<<8 | int(ip[3])
}

func intToIP(n int) net.IP {
	return net.IPv4(
		byte(n>>24),
		byte(n>>16),
		byte(n>>8),
		byte(n),
	)
}

func allocateIP(ctx context.Context, mac string) (string, error) {
	key := "dhcp:lease:" + mac

	// Check existing lease
	if ip, err := rdb.Get(ctx, key).Result(); err == nil {
		return ip, nil
	} else if err != redis.Nil {
		return "", fmt.Errorf("redis error: %v", err)
	}

	// Get next available IP
	ip, err := rdb.SPop(ctx, "dhcp:available_ips").Result()
	if err == redis.Nil {
		return "", fmt.Errorf("no available IPs in pool")
	} else if err != nil {
		return "", fmt.Errorf("failed to get IP from pool: %v", err)
	}

	// Store lease with 1-hour TTL
	if err := rdb.SetEx(ctx, key, ip, 3600).Err(); err != nil {
		// Return IP to pool if storage fails
		rdb.SAdd(ctx, "dhcp:available_ips", ip)
		return "", fmt.Errorf("failed to store lease: %v", err)
	}

	// Store reverse mapping
	rdb.SetEx(ctx, "dhcp:mac:"+ip, mac, 3600)

	return ip, nil
}

func releaseIP(ctx context.Context, ip string) error {
	// Get MAC for this IP
	mac, err := rdb.GetDel(ctx, "dhcp:mac:"+ip).Result()
	if err != nil {
		return fmt.Errorf("failed to get MAC for IP %s: %v", ip, err)
	}

	// Delete lease
	if err := rdb.Del(ctx, "dhcp:lease:"+mac).Err(); err != nil {
		return fmt.Errorf("failed to delete lease: %v", err)
	}

	// Return IP to pool
	if err := rdb.SAdd(ctx, "dhcp:available_ips", ip).Err(); err != nil {
		return fmt.Errorf("failed to return IP to pool: %v", err)
	}

	return nil
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Type        string `json:"type"`
			MAC         string `json:"mac"`
			RequestedIP string `json:"requested_ip,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		switch req.Type {
		case "DISCOVER":
			ip, err := allocateIP(ctx, req.MAC)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{
				"offer":      ip,
				"lease_time": 3600,
			})

		case "REQUEST":
			storedIP, err := rdb.Get(ctx, "dhcp:lease:"+req.MAC).Result()
			if err != nil || storedIP != req.RequestedIP {
				http.Error(w, "Invalid lease", http.StatusBadRequest)
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":     "ACK",
				"ip":         req.RequestedIP,
				"lease_time": 3600,
			})

		case "RELEASE":
			if err := releaseIP(ctx, req.RequestedIP); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)

		default:
			http.Error(w, "Invalid request type", http.StatusBadRequest)
		}
	})

	log.Println("Starting DHCP server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
