package main

import (
	"encoding/binary"
	"log"
	"net"
	"time"
)

// Lease information structure
type Lease struct {
	IP         net.IP
	SubnetMask net.IPMask
	Router     net.IP
	DNS        []net.IP
	LeaseTime  time.Duration
}

func main() {
	serverAddr := "255.255.255.255:67" // Broadcast address for DHCP server
	clientAddr := ":69"                // Local port for the client (DHCP clients use port 68)

	// Create a UDP connection
	conn, err := net.ListenPacket("udp4", clientAddr)
	if err != nil {
		log.Fatalf("Failed to create UDP connection: %v", err)
	}
	defer conn.Close()

	log.Printf("Client is running on %s and sending to server %s", conn.LocalAddr(), serverAddr)

	// Resolve the broadcast server address
	serverUDPAddr, err := net.ResolveUDPAddr("udp4", serverAddr)
	if err != nil {
		log.Fatalf("Failed to resolve server address: %v", err)
	}

	// Send DHCPDISCOVER and handle retries
	var offerPacket []byte
	for attempt := 1; attempt <= 3; attempt++ {
		log.Printf("Sending DHCPDISCOVER (Attempt %d)...", attempt)
		discoverPacket := createDHCPDiscoverPacket()
		_, err = conn.WriteTo(discoverPacket, serverUDPAddr)
		if err != nil {
			log.Fatalf("Failed to send DHCPDISCOVER: %v", err)
		}

		// Wait for DHCPOFFER
		offerPacket, err = receivePacket(conn)
		if err == nil {
			break
		}
		log.Printf("Retrying... (Error: %v)", err)
	}

	if err != nil {
		log.Fatalf("Failed to receive DHCPOFFER after 3 attempts: %v", err)
	}
	log.Println("DHCPOFFER received from server.")

	// Parse and validate the DHCPOFFER
	lease := parseDHCPOffer(offerPacket)

	// Send DHCPREQUEST to request the offered IP
	requestPacket := createDHCPRequestPacket(offerPacket, lease.IP)
	log.Println("Sending DHCPREQUEST to server...")
	_, err = conn.WriteTo(requestPacket, serverUDPAddr)
	if err != nil {
		log.Fatalf("Failed to send DHCPREQUEST: %v", err)
	}

	// Wait for DHCPACK
	ackPacket, err := receivePacket(conn)
	if err != nil {
		log.Fatalf("Failed to receive DHCPACK: %v", err)
	}
	log.Println("DHCPACK received from server.")

	// Parse and validate the DHCPACK
	lease = parseDHCPAck(ackPacket)
	log.Printf("Lease acquired: %+v", lease)

	// Simulate lease management
	log.Printf("Starting lease timer for %v...", lease.LeaseTime)
	time.Sleep(lease.LeaseTime)
	log.Println("Lease expired. Restarting DHCP process...")
}

// createDHCPDiscoverPacket creates a DHCPDISCOVER packet
func createDHCPDiscoverPacket() []byte {
	packet := make([]byte, 300)

	// DHCP header fields
	packet[0] = 0x01 // op: BOOTREQUEST
	packet[1] = 0x01 // htype: Ethernet
	packet[2] = 0x06 // hlen: MAC address length
	packet[3] = 0x00 // hops: 0
	// xid (transaction ID): A random number
	packet[4] = 0x39
	packet[5] = 0x03
	packet[6] = 0xF3
	packet[7] = 0x26
	// chaddr (client MAC address): A dummy MAC address
	copy(packet[28:34], []byte{0xDE, 0xAD, 0xBE, 0xEF, 0xCA, 0xFE})

	// Magic cookie to indicate DHCP protocol
	copy(packet[236:240], []byte{0x63, 0x82, 0x53, 0x63})

	// DHCP options
	packet[240] = 53  // Option: DHCP Message Type
	packet[241] = 1   // Length: 1 byte
	packet[242] = 1   // DHCPDISCOVER
	packet[243] = 255 // End option

	return packet
}

// receivePacket waits for a packet from the server with a timeout
func receivePacket(conn net.PacketConn) ([]byte, error) {
	buffer := make([]byte, 1500)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, addr, err := conn.ReadFrom(buffer)
	if err != nil {
		return nil, err
	}
	log.Printf("Received %d bytes from %s", n, addr)
	return buffer[:n], nil
}

// parseDHCPOffer parses a DHCPOFFER packet and extracts lease information
func parseDHCPOffer(packet []byte) Lease {
	offeredIP := net.IPv4(packet[16], packet[17], packet[18], packet[19])
	log.Printf("Offered IP address: %s", offeredIP)

	lease := Lease{IP: offeredIP}

	// Parse DHCP options (starting from byte 240)
	options := packet[240:]
	for i := 0; i < len(options); {
		if options[i] == 255 { // End option
			break
		}
		optionType := options[i]
		optionLen := options[i+1]
		optionData := options[i+2 : i+2+int(optionLen)]

		switch optionType {
		case 1: // Subnet mask
			lease.SubnetMask = net.IPMask(optionData)
		case 3: // Router
			lease.Router = net.IP(optionData)
		case 6: // DNS servers
			for j := 0; j < len(optionData); j += 10 {
				lease.DNS = append(lease.DNS, net.IP(optionData[j:j+4]))
			}
		case 51: // Lease time
			leaseTime := binary.BigEndian.Uint32(optionData)
			lease.LeaseTime = time.Duration(leaseTime) * time.Second
		}

		i += 2 + int(optionLen)
	}
	log.Printf("Parsed lease: %+v", lease)
	return lease
}

// createDHCPRequestPacket creates a DHCPREQUEST packet based on the DHCPOFFER
func createDHCPRequestPacket(offerPacket []byte, requestedIP net.IP) []byte {
	packet := make([]byte, 300)

	// Copy the initial structure from the offer packet
	copy(packet[:240], offerPacket[:240])

	// Set the DHCP message type to DHCPREQUEST
	copy(packet[236:240], []byte{0x63, 0x82, 0x53, 0x63})
	packet[240] = 53 // Option: DHCP Message Type
	packet[241] = 1  // Length: 1 byte
	packet[242] = 3  // DHCPREQUEST

	// Request the offered IP address
	packet[243] = 50 // Option: Requested IP Address
	packet[244] = 4  // Length: 4 bytes
	copy(packet[245:249], requestedIP.To4())

	packet[249] = 255 // End option

	return packet
}

// parseDHCPAck parses a DHCPACK packet and extracts lease information
func parseDHCPAck(packet []byte) Lease {
	return parseDHCPOffer(packet) // Reuse the DHCPOFFER parser as the structure is similar
}
