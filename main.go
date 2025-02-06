package main

import (
	"errors"
	"fmt"
	"net"
	"syscall"
	"time"

	"golang.org/x/net/ipv4"
	// "time"
)

const (
	MAX_HOPS = 3000
)

func TraceRoute(dest net.IP, timeout int) (ips []string, err error) {
	to := time.Millisecond * time.Duration(timeout)

	udpAddr := net.UDPAddr{IP: dest, Port: 33434}

	UDPconn, err := net.DialUDP("udp4", nil, &udpAddr)

	if err != nil {
		printErr(err, 1)
		return ips, err
	}

	UDPconn.SetDeadline(time.Now().Add(to))
	defer UDPconn.Close()

	for hop := 1; hop <= MAX_HOPS; hop++ {

		icmpConn, err := net.ListenPacket("ip4:icmp", "0.0.0.0")

		if err != nil {
			printErr(err, 2)
			return ips, err
		}

		// setting ttl
		ipv4.NewConn(UDPconn).SetTTL(hop)

		icmpConn.SetDeadline(time.Now().Add(to))

		defer icmpConn.Close()

		data := make([]byte, 0)

		_, err = UDPconn.Write(data)

		if err != nil {
			if errors.Is(err, syscall.EHOSTUNREACH) {
				return ips, nil
			}
			printErr(err, 3)
			return ips, err
		}

		// Create a buffer to read the response
		buffer := make([]byte, 11500)

		_, addr, err := icmpConn.ReadFrom(buffer)

		if err != nil {
			if errors.Is(err, syscall.ECONNREFUSED) {
				printErr(err, 4)

			} else if errors.Is(err, syscall.EHOSTUNREACH) {
				return ips, nil
			} else {
				printErr(err, 5)
			}
		}

		println(addr.String())

		ips = append(ips, addr.String())

		if dest.String() == addr.String() {
			return ips, err
		}

	}

	return ips, err

}

func printErr(error error, id int) {
	fmt.Printf("Error (%d): %s\n", id, error.Error())
}

func main() {
	ips, err := TraceRoute(net.IPv4(1, 1, 1, 1), 100)

	if err != nil {
		printErr(err, 888)
	}

	fmt.Printf("%v\n", ips)

}
