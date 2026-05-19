// Package scan provides types and functions to perform TCP port
// scans and a list of hosts
package scan

import (
	"fmt"
	"net"
	"time"
)

// Whether a port is open or closed
type state bool

// Convert the boolean value of state to a human readable string
func (s state) String() string {
	if s {
		return "open"
	}
	return "closed"
}

// PortState represents the state of a single TCP port
type PortState struct {
	Port int
	Open state
}

// scanPort performs a port scan on a single TCP port
func scanPort(host string, port int) PortState {
	p := PortState{
		Port: port,
	}

	address := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	scanConn, err := net.DialTimeout("tcp", address, 1*time.Second)

	if err != nil { // port is closed
		return p
	}

	scanConn.Close()
	p.Open = true
	return p
}

// Results represents the scan results for a single host
type Results struct {
	Host       string
	NotFound   bool        // host can be resolved to a valid IP address
	PortStates []PortState // status of each port scanned
}

// Run performs a port scan on the hosts list
func Run(hl *HostsList, ports []int) []Results {
	res := make([]Results, 0, len(ports))

	for _, h := range hl.Hosts {
		r := Results{
			Host: h,
		}

		if _, err := net.LookupHost(h); err != nil {
			r.NotFound = true
			res = append(res, r)
			continue
		}

		// Host found, execute port scan on each port
		for _, p := range ports {
			r.PortStates = append(r.PortStates, scanPort(h, p))
		}

		res = append(res, r)
	}

	return res
}
