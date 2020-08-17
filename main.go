package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"syscall"
	"time"
)

func getDialerForInterface(intf string) (*net.Dialer, error) {
	switch {
	case strings.Contains(intf, "ppp"):
		return getPPPDialer(intf), nil
	default:
		return getGenericDialer(intf)
	}
}

func getGenericDialer(intf string) (*net.Dialer, error) {
	ief, err := net.InterfaceByName(intf)
	if err != nil {
		return nil, fmt.Errorf("error getting '%s' interface by name: %s", intf, err)
	}

	addrs, err := ief.Addrs()
	if err != nil {
		return nil, fmt.Errorf("error getting '%s' addresses: %s", intf, err)
	}

	if len(addrs) < 1 {
		return nil, fmt.Errorf("interface '%s' had no IPs", intf)
	}

	tcpAddr := &net.TCPAddr{
		IP: addrs[0].(*net.IPNet).IP,
	}

	return &net.Dialer{
		Timeout:   30 * time.Second,
		LocalAddr: tcpAddr,
	}, nil
}

func getPPPDialer(intf string) *net.Dialer {
	return &net.Dialer{
		Timeout: 30 * time.Second,
		Control: func(network, address string, c syscall.RawConn) error {
			return c.Control(func(fd uintptr) {
				err := syscall.SetsockoptString(int(fd), syscall.SOL_SOCKET, 25, intf)
				if err != nil {
					log.Printf("control: %s", err)
					return
				}
			})
		}}
}

var help = `example:
try google.com on port 80 on the ppp0 interface
$ nettest ppp0 google.com 80
`

func main() {
	if len(os.Args) != 4 {
		fmt.Printf(help)
		log.Fatalf("wrong number of arguments")
	}

	// get the relevant args
	intf := os.Args[1]
	destAddr := os.Args[2]
	destPort := os.Args[3]

	// get the dialer based on the interface
	d, err := getDialerForInterface(intf)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := d.Dial("tcp", fmt.Sprintf("%s:%s", destAddr, destPort))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Successful connection with local IP %s on %s, to remote %s:%s\n", conn.LocalAddr().String(), intf, destAddr, destPort)

	fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")

	b, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", b)
}
