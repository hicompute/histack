package dhcp

import (
	"net"
	"sync"

	histack_ipam "github.com/hicompute/histack/pkg/ipam"
)

type DHCPV4Server struct {
	mu   sync.RWMutex
	ipam histack_ipam.IPAM
}

type DHCPV4Packet struct {
	op       byte             // Message op code (1=request, 2=reply)
	htype    byte             // Hardware address type (1=Ethernet)
	hlen     byte             // Hardware address length
	hops     byte             // Hops (used by relays)
	xid      uint32           // Transaction ID
	secs     uint16           // Seconds elapsed
	flags    uint16           // Flags
	ciaddr   net.IP           // Client IP
	yiaddr   net.IP           // Your (client) IP
	siaddr   net.IP           // Next server IP
	giaddr   net.IP           // Relay agent IP (CRITICAL FIELD)
	chaddr   net.HardwareAddr // Client MAC
	sname    []byte           // Server host name
	file     []byte           // Boot file name
	options  []byte           // DHCP options
	msgType  byte             // DHCP Message Type (Option 53)
	clientID []byte           // Client Identifier (Option 61)
}
