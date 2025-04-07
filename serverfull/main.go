package main

import (
	"log" // To log information and errors
	// For random number generation (used for selecting free lease IP)
	"net"  // For networking operations (working with IPs)
	"time" // For handling time and lease expiration

	dhcp "github.com/krolaw/dhcp4" // The dhcp4 package for DHCP protocol implementation
)

// lease represents a DHCP lease given to a client
// It contains the client's MAC address (nic) and the expiry time of the lease.
type lease struct {
	nic    string    // Client's MAC address (CHAddr) to identify the client
	expiry time.Time // Expiry time of the lease (when the lease should be considered invalid)
}


// DHCPHandler is the custom handler for DHCP server that implements the ServeDHCP method
// It manages the IP allocation, lease durations, and handles client requests.
type DHCPHandler struct {
	ip            net.IP        // The IP address of the DHCP server
	options       dhcp.Options  // DHCP options like subnet mask, router, DNS, etc.
	start         net.IP        // Starting IP address of the range of IPs to be distributed
	leaseRange    int           // Total number of IPs that the server can allocate
	leaseDuration time.Duration // Duration for which an IP lease is valid
	leases        map[int]lease // A map that keeps track of assigned leases (IP -> lease)
}

// ServeDHCP is the main method that handles the different types of DHCP messages:
// Discover, Request, and other types like Release, Decline.
func (h *DHCPHandler) ServeDHCP(p dhcp.Packet, msgType dhcp.MessageType, options dhcp.Options) (d dhcp.Packet) {
	switch msgType {

	// Handle DHCP Discover message (sent by the client to find a DHCP server)
	case dhcp.Discover:
		// Find a free lease IP for the client
		log.Println("Inside Discover")

		free := h.freeLease()
		if free == -1 {
			return nil // No available leases, so we don't respond
		}

		// Respond with a DHCP Offer message, providing the client with an IP and lease duration
		return dhcp.ReplyPacket(p, dhcp.Offer, h.ip, dhcp.IPAdd(h.start, free), h.leaseDuration,
			h.options.SelectOrderOrAll(options[dhcp.OptionParameterRequestList]))

	// Handle DHCP Request message (client requesting an IP after receiving a Discover message)
	case dhcp.Request:
		log.Println("Inside Request Packet")
		reqIP := net.IP(options[dhcp.OptionRequestedIPAddress]) // Requested IP by client
		if reqIP == nil {
			reqIP = net.IP(p.CIAddr()) // Fallback to client’s CIAddr (client IP address)
		}

		// Calculate the lease number based on the requested IP
		leaseNum := dhcp.IPRange(h.start, reqIP) - 1

		// Check if the requested IP is within the valid range of available IPs
		if leaseNum >= 0 && leaseNum < h.leaseRange {
			// Allocate the lease to the client (store it in the leases map with expiration time)
			h.leases[leaseNum] = lease{nic: p.CHAddr().String(), expiry: time.Now().Add(h.leaseDuration)}
			// Respond with a DHCP ACK, confirming the IP allocation
			return dhcp.ReplyPacket(p, dhcp.ACK, h.ip, reqIP, h.leaseDuration,
				h.options.SelectOrderOrAll(options[dhcp.OptionParameterRequestList]))
		}
		// If the requested IP is not valid, respond with a DHCP NAK (Negative Acknowledgement)
		return dhcp.ReplyPacket(p, dhcp.NAK, h.ip, nil, 0, nil)
	}

	return nil // Return nil for any unhandled message types
}

// freeLease looks for a free lease in the available IP range
// It randomly selects a lease and checks if it has expired. If expired or unassigned, it's marked free.
func (h *DHCPHandler) freeLease() int {
	now := time.Now() // Current time to check lease expiry
	for i := 0; i < h.leaseRange; i++ {
		// If the lease is not assigned or has expired, it's available for use
		if l, ok := h.leases[i]; !ok || l.expiry.Before(now) {
			return i // Return the available lease number (index)
		}
	}
	return -1 // No free leases available
}

// main function initializes and starts the DHCP server
func main() {
	// Define the server's IP address (this will be the IP the clients will get offers from)
	serverIP := net.IP{0,0,0,0} // {172, 30, 0, 1}

	// Create a new DHCPHandler with necessary configurations
	handler := &DHCPHandler{
		ip:            serverIP,                // The server's IP
		leaseDuration: 2 * time.Hour,           // Lease duration of 2 hours
		start:         net.IP{172, 25, 2, 2},   // Start IP of the range for allocation
		leaseRange:    5,                      // Allocate 50 IPs starting from the start IP
		leases:        make(map[int]lease, 10), // Initialize the leases map to track allocated IPs
		options: dhcp.Options{
			dhcp.OptionSubnetMask:       []byte{255, 255, 240, 0}, // Subnet mask
			dhcp.OptionRouter:           []byte(serverIP),         // Router is the same as the server IP
			dhcp.OptionDomainNameServer: []byte(serverIP),         // DNS is also the server
		},
	}

	// Log message indicating the server is starting
	log.Println("Starting DHCP server...")

	// Start the DHCP server and listen for incoming requests
	log.Fatal(dhcp.ListenAndServe(handler)) // This will block and run continuously
}
