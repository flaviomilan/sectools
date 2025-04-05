# 🛡️ sectools

This repository contains tools and experiments focused on offensive and defensive security, task automation, network analysis, and threat detection.

---

## 📁 Structure

Scripts are organized in folders, each dedicated to a specific functionality or technique.

---

## 📜 Available Scripts

### 🔐 `port-knocking-scanner`  

> 📂 `port-knocking-scanner/`

A network scanner that uses **port knocking** sequences to identify hosts running hidden services.  
Built using the [`gopacket`](https://github.com/google/gopacket) library to avoid external dependencies like `hping3`.

**Features:**
- 🔎 Scans a range of IPs within a `/24` network.
- 🛠️ Sends a configurable sequence of knock ports.
- 🔍 Checks if a service is exposed after the final knock.
- 🌐 Extracts HTTP response (if available) from the hidden service.
- 💡 Does not depend on external binaries like `hping3`.

**Usage:**

```bash
go build
sudo ./knocking_scanner -start 192.168.0.1 -end 192.168.0.254 -ports 13,37,30000,3000,1337 -iface eth0
```
