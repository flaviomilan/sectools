package main

import (
	"crypto/rand"
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/flaviomilan/sectools/libs/netutil"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

// version is set at build time via -ldflags.
var version = "dev"

func cryptoRandUint32() uint32 {
	var b [4]byte
	_, _ = rand.Read(b[:])
	return binary.LittleEndian.Uint32(b[:])
}

func sendSYN(ip string, port int, handle *pcap.Handle, srcPort layers.TCPPort, localIP net.IP) error {
	ipLayer := layers.IPv4{
		SrcIP:    localIP,
		DstIP:    net.ParseIP(ip),
		Protocol: layers.IPProtocolTCP,
	}
	tcp := layers.TCP{
		SrcPort: srcPort,
		DstPort: layers.TCPPort(uint16(port)), //nolint:gosec // port already validated in range 0-65535
		SYN:     true,
		Seq:     cryptoRandUint32(),
		Window:  14600,
	}
	if err := tcp.SetNetworkLayerForChecksum(&ipLayer); err != nil {
		return fmt.Errorf("set checksum: %w", err)
	}

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}
	if err := gopacket.SerializeLayers(buf, opts, &ipLayer, &tcp); err != nil {
		return fmt.Errorf("serialize: %w", err)
	}
	if err := handle.WritePacketData(buf.Bytes()); err != nil {
		return fmt.Errorf("write packet: %w", err)
	}
	return nil
}

func knockAndCheck(wg *sync.WaitGroup, ip string, ports []int, resultChan chan<- string, iface string, localIP net.IP) {
	defer wg.Done()

	handle, err := pcap.OpenLive(iface, 65536, false, pcap.BlockForever)
	if err != nil {
		fmt.Printf("[!] Error opening interface: %v\n", err)
		return
	}
	defer handle.Close()

	srcPort := layers.TCPPort(1024 + uint16(cryptoRandUint32()%(65535-1024))) //nolint:gosec // overflow impossible
	for _, port := range ports[:len(ports)-1] {
		if err := sendSYN(ip, port, handle, srcPort, localIP); err != nil {
			fmt.Printf("[!] Error sending SYN to %s:%d: %v\n", ip, port, err)
			return
		}
		time.Sleep(100 * time.Millisecond)
	}

	lastPort := ports[len(ports)-1]
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, lastPort), 1*time.Second)
	if err == nil {
		resultChan <- ip
		conn.Close()
	} else {
		fmt.Print(".")
	}
}

func main() {
	startIP := flag.String("start", "", "Start IP (e.g., 192.168.0.1)")
	endIP := flag.String("end", "", "End IP (e.g., 192.168.0.254)")
	portsInput := flag.String("ports", "13,37,30000,3000,1337", "Knock ports, comma-separated")
	iface := flag.String("iface", "eth0", "Network interface for packet injection")
	showVersion := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("port-knocking-scanner %s\n", version)
		return
	}

	if *startIP == "" || *endIP == "" {
		netutil.MustExitf("Usage: port-knocking-scanner -start <START_IP> -end <END_IP> [-ports <p1,p2,...>] -iface <iface>")
	}

	if !netutil.IsValidIP(*startIP) || !netutil.IsValidIP(*endIP) {
		netutil.MustExitf("[-] Invalid IP addresses")
	}

	net1 := strings.Join(strings.Split(*startIP, ".")[:3], ".")
	net2 := strings.Join(strings.Split(*endIP, ".")[:3], ".")

	if net1 != net2 {
		netutil.MustExitf("[-] IPs must be in the same /24 network")
	}

	ports, err := netutil.ParsePorts(*portsInput)
	if err != nil {
		netutil.MustExitf("[-] %v", err)
	}

	localIP, err := netutil.GetLocalIPv4(*iface)
	if err != nil {
		netutil.MustExitf("[-] %v", err)
	}

	start := netIPLastOctet(*startIP)
	end := netIPLastOctet(*endIP)

	var wg sync.WaitGroup
	resultChan := make(chan string, 256)

	fmt.Printf("[+] Scanning %s to %s on ports %v...\n", *startIP, *endIP, ports)

	for i := start; i <= end; i++ {
		ip := fmt.Sprintf("%s.%d", net1, i)
		wg.Add(1)
		go knockAndCheck(&wg, ip, ports, resultChan, *iface, localIP)
	}

	wg.Wait()
	close(resultChan)

	found := make([]string, 0, len(resultChan))
	for ip := range resultChan {
		found = append(found, ip)
	}

	fmt.Println()
	if len(found) > 0 {
		fmt.Printf("[+] Port knocking detected on %d host(s)!\n", len(found))
		for _, ip := range found {
			fmt.Printf("[+] Host: %s\n", ip)
		}
	} else {
		fmt.Println("[-] No hosts found.")
	}
}

func netIPLastOctet(ip string) int {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return 0
	}
	n, err := strconv.Atoi(parts[3])
	if err != nil {
		return 0
	}
	return n
}
