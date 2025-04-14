package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

func bannerGrab(host string, port string, timeout time.Duration, payload string) string {
	address := net.JoinHostPort(host, port)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return fmt.Sprintf("❌ Error connecting to %s: %v\n", address, err)
	}
	defer conn.Close()

	if payload != "" {
		_, err := conn.Write([]byte(payload))
		if err != nil {
			return fmt.Sprintf("⚠️ Failed to send data to %s: %v\n", address, err)
		}
	}

	conn.SetReadDeadline(time.Now().Add(timeout))
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		return fmt.Sprintf("⚠️ No banner received from %s (or connection closed): %v\n", address, err)
	}

	return fmt.Sprintf("✅ Banner from %s:\n%s\n", address, string(buffer[:n]))
}

func main() {
	host := flag.String("host", "", "Target host (e.g., scanme.nmap.org)")
	ports := flag.String("ports", "", "Target port(s), comma-separated (e.g., 22,80,443)")
	timeout := flag.Int("timeout", 5, "Connection timeout in seconds")
	payload := flag.String("send", "", "Optional payload to send before reading (e.g., GET / HTTP/1.0\\r\\n\\r\\n)")
	output := flag.String("output", "", "Optional output file to save banners")

	flag.Parse()

	if *host == "" || *ports == "" {
		fmt.Println("❗ Usage: go run main.go -host <host> -ports <port[,port,...]> [options]")
		flag.PrintDefaults()
		return
	}

	portList := strings.Split(*ports, ",")
	var results []string

	for _, port := range portList {
		port = strings.TrimSpace(port)
		result := bannerGrab(*host, port, time.Duration(*timeout)*time.Second, *payload)
		fmt.Print(result)
		results = append(results, result)
	}

	if *output != "" {
		err := os.WriteFile(*output, []byte(strings.Join(results, "\n")), 0644)
		if err != nil {
			fmt.Printf("❌ Error saving to file %s: %v\n", *output, err)
			return
		}
		fmt.Printf("📁 Results saved to: %s\n", *output)
	}
}
