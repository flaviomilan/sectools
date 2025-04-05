package main

import (
	"flag"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

func getLocalIP(interfaceName string) net.IP {
	ifaceObj, err := net.InterfaceByName(interfaceName)
	if err != nil {
		fmt.Printf("[-] Erro ao obter interface %s: %v\n", interfaceName, err)
		os.Exit(1)
	}

	addrs, err := ifaceObj.Addrs()
	if err != nil {
		fmt.Printf("[-] Erro ao obter enderecos da interface: %v\n", err)
		os.Exit(1)
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ip4 := ipNet.IP.To4(); ip4 != nil {
				return ip4
			}
		}
	}

	fmt.Println("[-] Nao foi possivel determinar o IP da interface.")
	os.Exit(1)
	return nil
}

func sendSYN(ip string, port int, handle *pcap.Handle, srcPort layers.TCPPort, localIP net.IP) {
	ipLayer := layers.IPv4{
		SrcIP:    localIP,
		DstIP:    net.ParseIP(ip),
		Protocol: layers.IPProtocolTCP,
	}
	tcp := layers.TCP{
		SrcPort: srcPort,
		DstPort: layers.TCPPort(port),
		SYN:     true,
		Seq:     rand.Uint32(),
		Window:  14600,
	}
	tcp.SetNetworkLayerForChecksum(&ipLayer)

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}
	gopacket.SerializeLayers(buf, opts, &ipLayer, &tcp)
	handle.WritePacketData(buf.Bytes())
}

func knockAndCheck(wg *sync.WaitGroup, ip string, ports []int, resultChan chan<- string, iface string, localIP net.IP) {
	defer wg.Done()

	handle, err := pcap.OpenLive(iface, 65536, false, pcap.BlockForever)
	if err != nil {
		fmt.Printf("Erro ao abrir interface: %v\n", err)
		return
	}
	defer handle.Close()

	srcPort := layers.TCPPort(rand.Intn(65535-1024) + 1024)
	for _, port := range ports[:len(ports)-1] {
		sendSYN(ip, port, handle, srcPort, localIP)
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
	startIP := flag.String("start", "", "IP inicial (ex: 192.168.0.1)")
	endIP := flag.String("end", "", "IP final (ex: 192.168.0.254)")
	portsInput := flag.String("ports", "13,37,30000,3000,1337", "Portas para knocking separadas por vírgula")
	iface := flag.String("iface", "eth0", "Interface de rede para envio dos pacotes")
	flag.Parse()

	if *startIP == "" || *endIP == "" {
		fmt.Println("[-] Uso: ./scanner -start <IP_INICIAL> -end <IP_FINAL> [-ports <porta1,porta2,...>] -iface <interface>")
		return
	}

	if !isValidIP(*startIP) || !isValidIP(*endIP) {
		fmt.Println("[-] Enderecos IP invalidos!")
		return
	}

	rede1 := strings.Join(strings.Split(*startIP, ".")[:3], ".")
	rede2 := strings.Join(strings.Split(*endIP, ".")[:3], ".")

	if rede1 != rede2 {
		fmt.Println("[-] Os IPs devem estar na mesma rede /24")
		return
	}

	inicio, _ := strconv.Atoi(strings.Split(*startIP, ".")[3])
	fim, _ := strconv.Atoi(strings.Split(*endIP, ".")[3])

	portsStr := strings.Split(*portsInput, ",")
	ports := []int{}
	for _, p := range portsStr {
		port, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			fmt.Printf("[-] Porta invalida: %s\n", p)
			return
		}
		ports = append(ports, port)
	}

	localIP := getLocalIP(*iface)

	var wg sync.WaitGroup
	resultChan := make(chan string, 256)

	fmt.Printf("[+] Testando hosts de %s a %s nas portas %v...\n", *startIP, *endIP, ports)

	for i := inicio; i <= fim; i++ {
		ip := fmt.Sprintf("%s.%d", rede1, i)
		wg.Add(1)
		go knockAndCheck(&wg, ip, ports, resultChan, *iface, localIP)
	}

	wg.Wait()
	close(resultChan)

	found := []string{}
	for ip := range resultChan {
		found = append(found, ip)
	}

	fmt.Println()
	if len(found) > 0 {
		fmt.Printf("[+] Port knocking encontrado em %d hosts!\n", len(found))
		for _, ip := range found {
			fmt.Printf("[+] Host identificado %s...\n", ip)
		}
	} else {
		fmt.Println("[-] Nenhum host comprometido encontrado.")
	}
}
