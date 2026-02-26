package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/flaviomilan/sectools/libs/netutil"
)

// version is set at build time via -ldflags.
var version = "dev"

func main() {
	host := flag.String("host", "", "Target host (e.g., scanme.nmap.org)")
	ports := flag.String("ports", "", "Target port(s), comma-separated (e.g., 22,80,443)")
	timeout := flag.Int("timeout", 5, "Connection timeout in seconds")
	payload := flag.String("send", "", "Optional payload to send before reading (e.g., GET / HTTP/1.0\\r\\n\\r\\n)")
	output := flag.String("output", "", "Optional output file to save banners")
	showVersion := flag.Bool("version", false, "Print version and exit")

	flag.Parse()

	if *showVersion {
		fmt.Printf("banner-grabber %s\n", version)
		return
	}

	if *host == "" || *ports == "" {
		fmt.Fprintln(os.Stderr, "Usage: banner-grabber -host <host> -ports <port[,port,...]> [options]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	portList := strings.Split(*ports, ",")
	dur := time.Duration(*timeout) * time.Second
	results := make([]string, 0, len(portList))

	for _, port := range portList {
		port = strings.TrimSpace(port)
		address := net.JoinHostPort(*host, port)

		banner, err := netutil.GrabBanner(*host, port, dur, *payload)
		if err != nil {
			msg := fmt.Sprintf("[!] %s: %v", address, err)
			fmt.Fprintln(os.Stderr, msg)
			results = append(results, msg)
			continue
		}

		msg := fmt.Sprintf("[+] %s\n%s", address, banner)
		fmt.Print(msg)
		results = append(results, msg)
	}

	if *output != "" {
		err := os.WriteFile(*output, []byte(strings.Join(results, "\n")), 0600)
		if err != nil {
			netutil.MustExitf("[!] Error saving to %s: %v", *output, err)
		}
		fmt.Fprintf(os.Stderr, "[*] Results saved to: %s\n", *output)
	}
}
