package dhcp

import (
	"encoding/binary"
	"net"

	histack_ipam "github.com/hicompute/histack/pkg/ipam"
	"k8s.io/klog/v2"
)

func Start() error {
	ipam := histack_ipam.New()
	dhcpV4Server := &DHCPV4Server{
		ipam: *ipam,
	}
	return dhcpV4Server.run()
}

func (d4s *DHCPV4Server) run() error {
	conn, err := net.ListenPacket("udp4", ":67")
	if err != nil {
		klog.Errorf("Failed to listen dhcp v4: %v", err)
		return err
	}
	buffer := make([]byte, 1500)
	for {
		n, addr, err := conn.ReadFrom(buffer)
		if err != nil {
			klog.Errorf("Read error: %v", err)
			continue
		}

		go d4s.handlePacket(conn, buffer[:n], addr)
	}
}

func (d4s *DHCPV4Server) handlePacket(conn net.PacketConn, data []byte, clientAddr net.Addr) {
	packet, err := d4s.parsePacket(data)
	if err != nil {
		return
	}
	klog.Infof("Received %v from MAC:%s via relay:%s", packet.msgType, packet.chaddr, packet.giaddr)
}

func (d4s *DHCPV4Server) parsePacket(data []byte) (*DHCPV4Packet, error) {
	if len(data) < 240 {
		return nil, net.InvalidAddrError("packet too short")
	}
	p := &DHCPV4Packet{
		op:     data[0],
		htype:  data[1],
		hlen:   data[2],
		hops:   data[3],
		xid:    binary.BigEndian.Uint32(data[4:8]),
		secs:   binary.BigEndian.Uint16(data[8:10]),
		flags:  binary.BigEndian.Uint16(data[10:12]),
		ciaddr: net.IP(data[12:16]),
		yiaddr: net.IP(data[16:20]),
		siaddr: net.IP(data[20:24]),
		giaddr: net.IP(data[24:28]),
	}
	if p.hlen <= 16 {
		p.chaddr = net.HardwareAddr(data[28 : 28+p.hlen])
	}

	return p, nil
}
