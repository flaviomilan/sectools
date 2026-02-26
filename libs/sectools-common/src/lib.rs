//! Shared utilities for sectools Rust tools.

use std::io::{Read, Write};
use std::net::{IpAddr, SocketAddr, TcpStream};
use std::time::Duration;

/// Check whether the given string is a valid IP address.
pub fn is_valid_ip(ip: &str) -> bool {
    ip.parse::<IpAddr>().is_ok()
}

/// Parse a comma-separated port string into a Vec of u16.
pub fn parse_ports(input: &str) -> Result<Vec<u16>, String> {
    input
        .split(',')
        .map(|s| {
            let s = s.trim();
            s.parse::<u16>()
                .map_err(|_| format!("invalid port: {s}"))
                .and_then(|p| {
                    if p == 0 {
                        Err("port 0 is not valid".to_string())
                    } else {
                        Ok(p)
                    }
                })
        })
        .collect()
}

/// Grab a banner from a TCP service.
pub fn grab_banner(
    host: &str,
    port: u16,
    timeout: Duration,
    payload: Option<&[u8]>,
) -> Result<String, String> {
    let addr: SocketAddr = format!("{host}:{port}")
        .parse()
        .map_err(|e| format!("invalid address: {e}"))?;

    let mut stream = TcpStream::connect_timeout(&addr, timeout)
        .map_err(|e| format!("connection failed: {e}"))?;

    stream
        .set_read_timeout(Some(timeout))
        .map_err(|e| format!("set timeout: {e}"))?;

    if let Some(data) = payload {
        stream
            .write_all(data)
            .map_err(|e| format!("send failed: {e}"))?;
    }

    let mut buf = vec![0u8; 4096];
    let n = stream
        .read(&mut buf)
        .map_err(|e| format!("read failed: {e}"))?;

    Ok(String::from_utf8_lossy(&buf[..n]).to_string())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_is_valid_ip() {
        assert!(is_valid_ip("192.168.1.1"));
        assert!(is_valid_ip("::1"));
        assert!(!is_valid_ip("not-an-ip"));
        assert!(!is_valid_ip(""));
    }

    #[test]
    fn test_parse_ports() {
        assert_eq!(parse_ports("80").unwrap(), vec![80]);
        assert_eq!(parse_ports("22,80,443").unwrap(), vec![22, 80, 443]);
        assert_eq!(parse_ports("22, 80, 443").unwrap(), vec![22, 80, 443]);
        assert!(parse_ports("abc").is_err());
        assert!(parse_ports("0").is_err());
    }
}
