package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
	_ "time/tzdata" // Add this line

	_ "github.com/lib/pq"
)

var DB *sql.DB

func handleUDPMessage(conn *net.UDPConn, buffer []byte, n int) {
	message := string(buffer[:n])

	if strings.Contains(message, "WAN155 is down") {
		fmt.Println("TRIGGER: WAN offline")
		notifyOfflineDetection("CLARO")
		persistData("CLARO", "offline")
	} else if strings.Contains(message, "WAN155 is up") {
		fmt.Println("TRIGGER: WAN online")
		notifyOnlineDetection("CLARO")
		persistData("CLARO", "online")
	} else if strings.Contains(message, "WAN455 is up") {
		fmt.Println("TRIGGER: USB Modem online")
		notifyOnlineDetection("USB Modem")
		persistData("USB Modem", "online")
	} else if strings.Contains(message, "WAN455 is down") {
		fmt.Println("TRIGGER: USB Modem offline")
		notifyOfflineDetection("USB Modem")
		persistData("USB Modem", "offline")
	} else if strings.Contains(message, "[WAN4] took effect") {
		fmt.Println("TRIGGER: Backup UP")
		backupTookEffect()
	} else {
		fmt.Println(message)
		fmt.Println("-----------------------------")
	}
}

func notifyOnlineDetection(wan string) {
	httpclient := &http.Client{}
	request, _ := http.NewRequest(
		"POST", "http://ntfy/network",
		strings.NewReader("Link Up on WAN ["+wan+"]"),
	)
	request.Header.Set("Tags", "+1,heavy_check_mark")
	do, err := httpclient.Do(request)
	if err != nil {
		fmt.Println("Error ,", err.Error())
	}
	fmt.Println("NTFY response: ", do.StatusCode)
}

func backupTookEffect() {
	httpclient := &http.Client{}
	request, _ := http.NewRequest(
		"POST", "http://ntfy/network",
		strings.NewReader("Backup Link took effect"),
	)
	request.Header.Set("Tags", "warning,skull")
	do, err := httpclient.Do(request)
	if err != nil {
		fmt.Println("Error ", err.Error())
	}
	fmt.Println("NTFY response: ", do.StatusCode)
}

func notifyOfflineDetection(wan string) {
	httpclient := &http.Client{}
	request, _ := http.NewRequest(
		"POST", "http://ntfy/network", strings.NewReader("Link Down on WAN ["+wan+"]"),
	)
	request.Header.Set("Tags", "warning,skull")
	do, err := httpclient.Do(request)
	if err != nil {
		fmt.Println("Error ", err.Error())
	}
	fmt.Println("NTFY response: ", do.StatusCode)
}

func main() {
	fmt.Println("Starting Syslog server")
	addr, err := net.ResolveUDPAddr("udp", ":514")
	if err != nil {
		log.Fatal(err)
	}

	setDataBase()

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

func persistData(wan, status string) {
	location, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		log.Fatal(err)
	}
	saoPauloTime := time.Now().In(location)
	exec, err := DB.Exec(
		"INSERT INTO network_status (wan, status, create_at) VALUES ($1, $2, $3)", wan, status,
		saoPauloTime,
	)
	if err != nil {
		fmt.Println("Error inserting data:", err)
	}
	rows, _ := exec.RowsAffected()
	if rows > 0 {
		fmt.Println("Data inserted successfully")
	} else {
		fmt.Println("Failed to insert data")
	}
}

func setDataBase() {
	const (
		host   = "postgresql"
		port   = 5432
		user   = "postgres"
		dbname = "postgres"
	)

	psqlInfo := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, os.Getenv("DB_PASSWORD"), dbname,
	)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}

	DB = db
}
