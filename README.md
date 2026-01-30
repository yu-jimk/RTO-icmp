# RTO-icmp

A simple CLI tool that sends ICMP Echo Requests and calculates the optimal Retransmission Timeout (RTO) based on RFC 6298

## Usage

### Prerequisites

- Go (1.18+ recommended)
- Root privileges (or capabilities) to send ICMP via raw sockets on most systems

### Build

Run the following in the project root to build the executable:

```bash
go build ./cmd
```

### Run

- The tool sends 10 ICMP Echo Requests and updates RTO according to RFC 6298.
- By default the tool targets `8.8.8.8`. Use the `-t` flag to set a different target (IP or hostname).
- Because the program uses raw sockets, you will likely need to run it with elevated privileges (e.g. `sudo`) on macOS/Linux.

Examples:

```bash
# Run with default target (may require sudo)
sudo ./main

# Run against example.com with a custom target string
sudo ./main -t example.com
```

### Notes

- `-t` : target IP address or hostname (default: `8.8.8.8`).
- If you prefer not to run as root, configure the binary with appropriate capabilities (Linux) or run in an environment that permits raw socket access.

If you need help or want to extend behavior (e.g., change the number of pings or output format), see the source under `cmd/` and `pkg/`.
