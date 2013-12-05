// Forked by Bence Tamas <mr.bence.tamas@gmail.com>

// The original package is written by Hailiang Wang. 
// https://github.com/hailiang/gosocks
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package socks implements a SOCKS5 proxy client.

A complete example using this package:
	package main

	import (
		"github.com/benctamas/gosocks"
		"fmt"
		"net/http"
		"io/ioutil"
	)

	func main() {
		dialSocks5Proxy := socks.DialSocks5Proxy("127.0.0.1:1080")
		tr := &http.Transport{Dial: dialSocks5Proxy}
		httpClient := &http.Client{Transport: tr}

		bodyText, err := TestHttpsGet(httpClient, "https://github.com/about")
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Print(bodyText)
	}

	func TestHttpsGet(c *http.Client, url string) (bodyText string, err error) {
		resp, err := c.Get(url)
		if err != nil { return }
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil { return }
		bodyText = string(body)
		return
	}
*/
package socks

import (
	"errors"
	"fmt"
	"net"
	"strconv"
)

// DialSocks5Proxy returns the dial function to be used in http.Transport object.
// Argument proxy should be in this format "127.0.0.1:1080".
func DialSocks5Proxy(proxy string) func(string, string) (net.Conn, error) {
	return func(_, targetAddr string) (conn net.Conn, err error) {
			return dialSocks5(proxy, targetAddr)
	}
}

func dialSocks5(proxy, targetAddr string) (conn net.Conn, err error) {
	// dial TCP
	conn, err = net.Dial("tcp", proxy)
	if err != nil {
		return
	}

	// version identifier/method selection request
	req := []byte{
		5, // version number
		1, // number of methods
		0, // method 0: no authentication (only anonymous access supported for now)
	}
	resp, err := sendReceive(conn, req)
	if err != nil {
		return
	} else if len(resp) != 2 {
		err = errors.New("Server does not respond properly.")
	} else if resp[0] != 5 {
		err = errors.New("Server does not support Socks 5.")
	} else if resp[1] != 0 { // no auth
		err = errors.New("socks method negotiation failed.")
		return
	}
	// detail request
	host, port, err := splitHostPort(targetAddr)
	if err != nil {
		return
	}
	ip, err := LookupIP4(host)
	if err != nil {
		return
	}
	req = []byte{
		5,              // version number
		1,              // connect command
		0,              // reserved, must be zero
		1,              // 1: ipv4 address
		ip[0], ip[1], ip[2], ip[3],
	}
	req = append(req, []byte{
		byte(port >> 8), // higher byte of destination port
		byte(port),      // lower byte of destination port (big endian)
	}...)
	resp, err = sendReceive(conn, req)
	if err != nil {
		return
	} else if len(resp) != 10 {
		err = errors.New("Server does not respond properly.")
	} else if resp[1] != 0 {
		err = errors.New("Can't complete SOCKS5 connection.")
	}

	return
}

func sendReceive(conn net.Conn, req []byte) (resp []byte, err error) {
	_, err = conn.Write(req)
	if err != nil {
		return
	}
	resp, err = readAll(conn)
	return
}

func readAll(conn net.Conn) (resp []byte, err error) {
	resp = make([]byte, 1024)
	n, err := conn.Read(resp)
	resp = resp[:n]
	return
}

func LookupIP4(host string) (ip net.IP, err error) {
	ips, err := net.LookupIP(host)
	if err != nil {
		return
	}
	if len(ips) == 0 {
		err = errors.New(fmt.Sprintf("Cannot resolve host: %s.", host))
		return
	}
	for _, ip = range ips {
		ip = ip.To4()
		if ip != nil {
			return
		}
	}
	err = errors.New(fmt.Sprintf("Cannot resolve IPv4 address of the host: %s.", host))
	return
}

func splitHostPort(addr string) (host string, port uint16, err error) {
	host, portStr, err := net.SplitHostPort(addr)
	portInt, err := strconv.ParseUint(portStr, 10, 16)
	port = uint16(portInt)
	return
}
