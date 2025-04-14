# 🛡️ sectools

This repository contains tools and experiments focused on offensive and defensive security, task automation, network analysis, and threat detection.

## 📜 Available Scripts

### 🔐 `port-knocking-scanner`

A network scanner that uses **port knocking** sequences to identify hosts running hidden services.  
Built using the [`gopacket`](https://github.com/google/gopacket) library to avoid external dependencies like `hping3`.

**Features:**

- 🔎 Scans a range of IPs within a `/24` network.
- 🛠️ Sends a configurable sequence of knock ports.
- 🔍 Checks if a service is exposed after the final knock.

**Usage:**

```bash
go build
sudo ./knocking_scanner -start 192.168.0.1 -end 192.168.0.254 -ports 13,37,30000,3000,1337 -iface eth0
```

### 🚩 `banner-grabber`

A tool to perform banner grabbing on specified hosts and ports, retrieving service information.
It allows TCP and UDP scanning and includes customizable timeout settings.

**Features:**

- 🎯 Target specific hosts and ports.
- 📜 Retrieve service banners.
- ⏱️ Customizable timeout.
- 🌐 Supports both TCP and UDP protocols.

**Usage:**

```bash
go build

./banner-grabber -host 192.168.1.10 -ports 80
./banner-grabber -host 192.168.1.10 -ports 22 -timeout 5
./banner-grabber -host 192.168.1.10 -ports 161 -udp
```
