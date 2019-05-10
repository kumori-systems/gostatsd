// Package udpclient is just to send data via udp
package udpclient

import (
	"errors"
	"fmt"
	"net"
	"strconv"
)

// UDPClient is a wrapper to send data via udp to an endpoint
type UDPClient struct {
	endpoint *net.UDPAddr
}

// New creates a new UDPClient
func New(host string, port int) (client *UDPClient, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("UDPClient.New: %s", r.(error).Error())
		}
	}()
	endpoint, err := net.ResolveUDPAddr("udp", host+":"+strconv.FormatInt(int64(port), 10))
	if err == nil {
		client = &UDPClient{endpoint}
	}
	return
}

// Send a byte array via udp
func (client *UDPClient) Send(data []byte) (err error) {
	var conn net.Conn
	conn = nil
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("udpclient: %s", r.(error).Error())
		}
		if conn != nil {
			conn.Close()
		}
	}()
	if client.endpoint == nil {
		err = errors.New("UDPClient.Send: endpoint undefined")
		return
	}
	conn, err = net.DialUDP("udp", nil, client.endpoint)
	if err != nil {
		err = fmt.Errorf("UDPClient.Send: %s", err.Error())
		return
	}
	var sentBytes int
	sentBytes, err = conn.Write(data)
	fmt.Println("JJJ --- ", client.endpoint, sentBytes, len(data))
	if err != nil {
		err = fmt.Errorf("UDPClient.Send: %s", err.Error())
		return
	}
	if sentBytes != len(data) {
		err = fmt.Errorf("UDPClient.Send: sent %d bytes, expected %d", sentBytes, len(data))
		return
	}
	return
}
