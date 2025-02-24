package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"syscall"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const (
	MAX_HOPS = 30
)

type Result struct {
	ip   net.Addr
	ping time.Duration
	hop  int
}

func ICMPTracert(dest string, timeout int) (ips []Result, err error) {
	to := time.Second * time.Duration(timeout)

	address, err := net.ResolveIPAddr("ip4", dest)

	if err != nil {
		printErr(err, 1)
		return ips, err
	}

	icmpConn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")

	if err != nil {
		printErr(err, 2)
		return ips, err
	}

	icmpConn.SetDeadline(time.Now().Add(to))

	defer icmpConn.Close()

	for ttl := 1; ttl <= MAX_HOPS; ttl++ {
		// setting ttl
		icmpConn.IPv4PacketConn().SetTTL(ttl)

		sendMsg := icmp.Message{
			Type: ipv4.ICMPTypeEcho,
			Code: 0,
			Body: &icmp.Echo{
				ID:   os.Getpid() & 0xffff,
				Seq:  ttl,
				Data: []byte(""),
			},
		}

		writeBuf, err := sendMsg.Marshal(nil)

		if err != nil {
			printErr(err, 2)
			return ips, err
		}

		start := time.Now()

		if _, err := icmpConn.WriteTo(writeBuf, address); err != nil {
			if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
				ips = append(ips, Result{})
				fmt.Println("*")
				if ttl == MAX_HOPS {
					return ips, err
				}
				continue
			} else {
				printErr(err, 4)
				return ips, err
			}
		}

		buffer := make([]byte, 11500)

		n, addr, err := icmpConn.ReadFrom(buffer)

		if err != nil {
			if errors.Is(err, syscall.EHOSTUNREACH) {
				return ips, nil
			}

			if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
				ips = append(ips, Result{})
				fmt.Println("*")
				if ttl == MAX_HOPS {
					return ips, err
				}
				continue
			} else {
				printErr(err, 4)
				return ips, err
			}

		}

		duration := time.Since(start)

		incomingMessage, err := icmp.ParseMessage(1, buffer[:n])

		if err != nil {
			return nil, err
		}

		switch incomingMessage.Type {
		case ipv4.ICMPTypeEchoReply:
			res := Result{
				ip:   addr,
				ping: duration,
				hop:  ttl,
			}
			fmt.Printf("%d %s (%v)\n", ttl, addr, duration)
			ips = append(ips, res)
			return ips, nil
		case ipv4.ICMPTypeTimeExceeded:
			res := Result{
				ip:   addr,
				ping: duration,
				hop:  ttl,
			}
			fmt.Printf("%d %s (%v)\n", ttl, addr, duration)
			ips = append(ips, res)
		default:
			ips = append(ips, Result{})
		}
	}

	return ips, err

}

func printErr(error error, id int) {
	fmt.Printf("Error (%d): %s\n", id, error.Error())
}

func main() {
	ips, err := ICMPTracert("177.79.86.166", 10)

	if err != nil {
		printErr(err, 0)
	}

	fmt.Printf("%v\n", &ips)

}
