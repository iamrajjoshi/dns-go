package main

import (
	"encoding/binary"
	"fmt"
	"net"
)

type Message struct {
	Header Header
}

type Header struct {
	/*
	   	                              1  1  1  1  1  1
	   	0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
	   +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	   |                      ID                       |
	   +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	   |QR|   Opcode  |AA|TC|RD|RA|   Z    |   RCODE   |
	   +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	   |                    QDCOUNT                    |
	   +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	   |                    ANCOUNT                    |
	   +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	   |                    NSCOUNT                    |
	   +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	   |                    ARCOUNT                    |
	   +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	*/
	ID      uint16 // 16 bits
	FLAGS   uint16 // 16 bits
	QDCOUNT uint16 // 16 bits
	ANCOUNT uint16 // 16 bits
	NSCOUNT uint16 // 16 bits
	ARCOUNT uint16 // 16 bits
}

func buildNewHeader() *Header {
	return &Header{
		ID:      1234,
		FLAGS:   0x8000,
		QDCOUNT: 0,
		ANCOUNT: 0,
		NSCOUNT: 0,
		ARCOUNT: 0,
	}
}

func (h *Header) toBytes() []byte {
	buf := make([]byte, 12)

	binary.BigEndian.PutUint16(buf[0:2], h.ID)
	binary.BigEndian.PutUint16(buf[2:4], h.FLAGS)
	binary.BigEndian.PutUint16(buf[4:6], h.QDCOUNT)
	binary.BigEndian.PutUint16(buf[6:8], h.ANCOUNT)
	binary.BigEndian.PutUint16(buf[8:10], h.NSCOUNT)
	binary.BigEndian.PutUint16(buf[10:12], h.ARCOUNT)
	return buf
}

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:2053")
	if err != nil {
		fmt.Println("Failed to resolve UDP address:", err)
		return
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("Failed to bind to address:", err)
		return
	}
	defer udpConn.Close()

	buf := make([]byte, 512)

	for {
		size, source, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error receiving data:", err)
			break
		}

		receivedData := string(buf[:size])
		fmt.Printf("Received %d bytes from %s: %s\n", size, source, receivedData)

		header := buildNewHeader()
		response := header.toBytes()

		_, err = udpConn.WriteToUDP(response, source)
		if err != nil {
			fmt.Println("Failed to send response:", err)
		}
	}
}
