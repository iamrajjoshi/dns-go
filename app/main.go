package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
)

type Message struct {
	Header   Header
	Question Question
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

type TYPE uint16
type CLASS uint16

const (
	TYPE_A     TYPE = 1
	TYPE_NS    TYPE = 2
	TYPE_MD    TYPE = 3
	TYPE_MF    TYPE = 4
	TYPE_CNAME TYPE = 5
	TYPE_SOA   TYPE = 6
	TYPE_MB    TYPE = 7
	TYPE_MG    TYPE = 8
	TYPE_MR    TYPE = 9
	TYPE_NULL  TYPE = 10
	TYPE_WKS   TYPE = 11
	TYPE_PTR   TYPE = 12
	TYPE_HINFO TYPE = 13
	TYPE_MINFO TYPE = 14
	TYPE_MX    TYPE = 15
	TYPE_TXT   TYPE = 16
)

const (
	CLASS_IN CLASS = 1
	CLASS_CS CLASS = 2
	CLASS_CH CLASS = 3
	CLASS_HS CLASS = 4
)

type Question struct {
	Name  []byte
	Type  TYPE
	Class CLASS
}

func buildNewQuestion() *Question {
	return &Question{
		Name:  []byte{},
		Type:  TYPE_A,
		Class: CLASS_IN,
	}
}

func (m *Message) encodeDomains(domains []string) {
	for _, domain := range domains {
		labels := strings.Split(domain, ".")
		for _, label := range labels {
			m.Question.Name = append(m.Question.Name, byte(len(label)))
			m.Question.Name = append(m.Question.Name, label...)
		}
	}
	m.Question.Name = append(m.Question.Name, '\x00')
	m.Header.QDCOUNT = uint16(len(domains))
}

func (m *Message) toBytes() []byte {
	buf := make([]byte, 0)

	buf = append(buf, m.Header.toBytes()...)
	buf = append(buf, m.Question.toBytes()...)

	return buf
}

func (q *Question) toBytes() []byte {
	buf := make([]byte, 4+len(q.Name))

	copy(buf[0:], q.Name)
	binary.BigEndian.PutUint16(buf[len(q.Name):len(q.Name)+2], uint16(q.Type))
	binary.BigEndian.PutUint16(buf[len(q.Name)+2:len(q.Name)+4], uint16(q.Class))

	return buf
}

func testing() []byte {
	header := buildNewHeader()
	question := buildNewQuestion()
	message := Message{
		Header:   *header,
		Question: *question,
	}
	message.encodeDomains([]string{"codecrafters.io"})
	return message.toBytes()
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

		response := testing()

		_, err = udpConn.WriteToUDP(response, source)
		if err != nil {
			fmt.Println("Failed to send response:", err)
		}
	}
}
