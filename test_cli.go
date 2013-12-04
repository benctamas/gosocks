package main

import (
	"gosocks"
	"fmt"
	"net/http"
	"io/ioutil"
	"flag"
)

var proxy_addr = flag.String("proxy", "localhost:1080", "proxy_ip_or_domain:port")
var target_url = flag.String("url", "http://github.com/about/", "url")

func main() {
    flag.Parse()
	dialSocks5Proxy := socks.DialSocks5Proxy(*proxy_addr)
	tr := &http.Transport{Dial: dialSocks5Proxy}
	httpClient := &http.Client{Transport: tr}

	bodyText, err := TestHttpsGet(httpClient, *target_url)
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
