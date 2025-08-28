package mpstats

import (
	"net"
	"net/http"
	"time"
)

const baseURL = "https://mpstats.io/api/wb/get"

type Client struct {
	token string
	http  *http.Client
}

func New(token string) *Client {
	tr := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	return &Client{
		token: token,
		http:  &http.Client{Timeout: 30 * time.Second, Transport: tr},
	}
}
