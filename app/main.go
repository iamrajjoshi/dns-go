package main

import (
	"fmt"
	"net"
)

func processMessage(buf []byte) []byte {
	// header := buildNewHeader()
	// question := buildNewQuestion()
	// answer := buildNewAnswer()

	// message := Message{
	// 	Header:   *header,
	// 	Question: *question,
	// 	Answer:   *answer,
	// }
	message := buildResponse(fromBytes(buf))

	message.Question.Name = encodeDomains([]string{"codecrafters.io"})

	message.Answer.Name = encodeDomains([]string{"codecrafters.io"})
	message.Answer.RDLENGTH = 4
	message.Answer.RDATA = []byte{127, 0, 0, 1}

	message.Header.QDCOUNT = uint16(1)
	message.Header.ANCOUNT = uint16(1)

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

		response := processMessage(buf[:size])

		_, err = udpConn.WriteToUDP(response, source)
		if err != nil {
			fmt.Println("Failed to send response:", err)
		}
	}
}
