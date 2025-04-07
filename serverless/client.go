package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	knativeURL := "http://dhcp-service.default.svc.cluster.local"
	mac := "de:ad:be:ef:ca:fe"

	// 1. DHCPDISCOVER (HTTP)
	discover := map[string]string{"type": "DISCOVER", "mac": mac}
	offer := sendDHCPRequest(knativeURL, discover)
	log.Printf("OFFER: %+v", offer)

	// 2. DHCPREQUEST (HTTP)
	request := map[string]string{"type": "REQUEST", "mac": mac, "ip": offer["ip"]}
	ack := sendDHCPRequest(knativeURL, request)
	log.Printf("ACK: %+v", ack)
}

func sendDHCPRequest(url string, req map[string]string) map[string]string {
	body, _ := json.Marshal(req)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Fatal("DHCP request failed:", err)
	}
	defer resp.Body.Close()

	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)
	return result
}
