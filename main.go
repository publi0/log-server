package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"
)

func handleUDPMessage(conn *net.UDPConn, buffer []byte, n int) {
	message := string(buffer[:n])
	// Define a regular expression to extract timestamp, device name, severity, IP address, and action
	re := regexp.MustCompile(`<(\d+)>(\w{3} \d{1,2} \d{2}:\d{2}:\d{2}) (\w+): (.+?): (.+)`)
	matches := re.FindStringSubmatch(message)

	if len(matches) < 6 {
		fmt.Println("Invalid log message format:", message)
		return
	}

	severity := matches[1]
	timestamp := matches[2]
	device := matches[3]
	ip := matches[4]
	action := matches[5]

	httpclient := &http.Client{}

	if strings.Contains(action, "is down.") {
		request, _ := http.NewRequest(
			"POST", "http://ntfy.home/home_infra",
			strings.NewReader(action),
		)
		request.Header.Set("Title", "Link Down")
		request.Header.Set("Tags", "warning,skull")
		httpclient.Do(request)
	} else if strings.Contains(action, "is up.") {
		request, _ := http.NewRequest(
			"POST", "http://ntfy.home/home_infra",
			strings.NewReader(action),
		)
		request.Header.Set("Title", "Link Up")
		request.Header.Set("Tags", "+1,heavy_check_mark")
		httpclient.Do(request)
	} else {
		fmt.Printf("Timestamp: %s\n", timestamp)
		fmt.Printf("Device: %s\n", device)
		fmt.Printf("Severity: %s\n", severity)
		fmt.Printf("IP Address: %s\n", ip)
		fmt.Printf("Action: %s\n", action)
		fmt.Println("-----------------------------")
	}
}

func main() {
	addr, err := net.ResolveUDPAddr("udp", ":514")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	fmt.Println("Syslog server is listening on UDP port 514")

	for {
		buffer := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Println("Error reading from connection:", err)
			continue
		}
		go handleUDPMessage(conn, buffer, n)
	}
}
