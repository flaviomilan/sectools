//! Fast asynchronous TCP port scanner for hosts and subnets.

use std::net::{IpAddr, Ipv4Addr};
use std::sync::Arc;
use std::time::Duration;

use clap::Parser;
use sectools_common::{is_valid_ip, parse_ports};
use tokio::io::AsyncWriteExt;
use tokio::net::TcpStream;
use tokio::sync::Semaphore;
use tokio::time::timeout;

/// Fast asynchronous TCP port scanner for hosts and subnets.
#[derive(Parser)]
#[command(name = "subnet-scanner", version, about)]
struct Cli {
    /// Target host or CIDR (e.g. 192.168.1.1 or 192.168.1.0/24)
    #[arg(short, long)]
    target: String,

    /// Comma-separated ports to scan (e.g. 22,80,443)
    #[arg(
        short,
        long,
        default_value = "21,22,23,25,53,80,110,143,443,993,995,3306,3389,5432,8080,8443"
    )]
    ports: String,

    /// Connection timeout in milliseconds
    #[arg(long, default_value = "1000")]
    timeout: u64,

    /// Maximum concurrent connections
    #[arg(short, long, default_value = "500")]
    concurrency: usize,

    /// Output file path (default: stdout)
    #[arg(short, long)]
    output: Option<String>,
}

/// Expand a target string (single IP or CIDR) into a list of host addresses.
fn expand_cidr(cidr: &str) -> Result<Vec<IpAddr>, String> {
    if let Some((ip_str, prefix_str)) = cidr.split_once('/') {
        let ip: IpAddr = ip_str.parse().map_err(|e| format!("invalid IP: {e}"))?;
        let prefix: u32 = prefix_str
            .parse()
            .map_err(|e| format!("invalid prefix: {e}"))?;

        match ip {
            IpAddr::V4(v4) => {
                if prefix > 32 {
                    return Err("prefix must be <= 32".to_string());
                }
                if prefix < 16 {
                    return Err("prefix must be >= 16 to avoid scanning too many hosts".to_string());
                }

                let base = u32::from(v4);
                let mask = if prefix == 32 {
                    u32::MAX
                } else {
                    !((1u32 << (32 - prefix)) - 1)
                };
                let network = base & mask;
                let broadcast = network | !mask;

                if prefix == 32 {
                    Ok(vec![IpAddr::V4(v4)])
                } else {
                    let mut addrs = Vec::with_capacity((broadcast - network - 1) as usize);
                    for i in (network + 1)..broadcast {
                        addrs.push(IpAddr::V4(Ipv4Addr::from(i)));
                    }
                    Ok(addrs)
                }
            }
            IpAddr::V6(_) => Err("IPv6 CIDR scanning is not supported yet".to_string()),
        }
    } else {
        if !is_valid_ip(cidr) {
            return Err(format!("invalid target: {cidr}"));
        }
        let ip: IpAddr = cidr.parse().map_err(|e| format!("invalid IP: {e}"))?;
        Ok(vec![ip])
    }
}

#[tokio::main]
async fn main() {
    let cli = Cli::parse();

    let hosts = match expand_cidr(&cli.target) {
        Ok(h) => h,
        Err(e) => {
            eprintln!("Error: {e}");
            std::process::exit(1);
        }
    };

    let ports = match parse_ports(&cli.ports) {
        Ok(p) => p,
        Err(e) => {
            eprintln!("Error: {e}");
            std::process::exit(1);
        }
    };

    let total = hosts.len() * ports.len();
    eprintln!(
        "Scanning {} host(s) x {} port(s) = {} probes (concurrency: {})",
        hosts.len(),
        ports.len(),
        total,
        cli.concurrency
    );

    let timeout_dur = Duration::from_millis(cli.timeout);
    let sem = Arc::new(Semaphore::new(cli.concurrency));
    let (tx, mut rx) = tokio::sync::mpsc::channel::<String>(1024);

    for host in &hosts {
        for &port in &ports {
            let sem = Arc::clone(&sem);
            let tx = tx.clone();
            let host = *host;

            tokio::spawn(async move {
                let _permit = sem.acquire().await.expect("semaphore closed");
                let addr = format!("{host}:{port}");
                if let Ok(Ok(_)) = timeout(timeout_dur, TcpStream::connect(&addr)).await {
                    let _ = tx.send(format!("{host}:{port} open")).await;
                }
            });
        }
    }

    drop(tx);

    let mut results = Vec::new();
    while let Some(line) = rx.recv().await {
        results.push(line);
    }
    results.sort();

    if results.is_empty() {
        eprintln!("No open ports found.");
        return;
    }

    let output_text = results.join("\n");

    match &cli.output {
        Some(path) => {
            let mut file = tokio::fs::File::create(path)
                .await
                .expect("cannot create output file");
            file.write_all(output_text.as_bytes())
                .await
                .expect("write failed");
            file.write_all(b"\n").await.expect("write failed");
            eprintln!("Results written to {path}");
        }
        None => {
            println!("{output_text}");
        }
    }

    eprintln!("Scan complete: {} open port(s) found.", results.len());
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_expand_cidr_single_host() {
        let hosts = expand_cidr("192.168.1.1").unwrap();
        assert_eq!(hosts.len(), 1);
        assert_eq!(hosts[0].to_string(), "192.168.1.1");
    }

    #[test]
    fn test_expand_cidr_slash_32() {
        let hosts = expand_cidr("10.0.0.5/32").unwrap();
        assert_eq!(hosts.len(), 1);
        assert_eq!(hosts[0].to_string(), "10.0.0.5");
    }

    #[test]
    fn test_expand_cidr_slash_24() {
        let hosts = expand_cidr("192.168.1.0/24").unwrap();
        assert_eq!(hosts.len(), 254);
        assert_eq!(hosts[0].to_string(), "192.168.1.1");
        assert_eq!(hosts[253].to_string(), "192.168.1.254");
    }

    #[test]
    fn test_expand_cidr_slash_30() {
        let hosts = expand_cidr("10.0.0.0/30").unwrap();
        assert_eq!(hosts.len(), 2);
        assert_eq!(hosts[0].to_string(), "10.0.0.1");
        assert_eq!(hosts[1].to_string(), "10.0.0.2");
    }

    #[test]
    fn test_expand_cidr_invalid_ip() {
        assert!(expand_cidr("not-an-ip").is_err());
    }

    #[test]
    fn test_expand_cidr_prefix_too_large() {
        assert!(expand_cidr("10.0.0.0/33").is_err());
    }

    #[test]
    fn test_expand_cidr_prefix_too_small() {
        assert!(expand_cidr("10.0.0.0/15").is_err());
    }

    #[test]
    fn test_expand_cidr_ipv6_unsupported() {
        assert!(expand_cidr("::1/128").is_err());
    }
}
