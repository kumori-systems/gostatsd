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
	endpoint string
}

// New creates a new UDPClient
func New(host string, port int) (client UDPClient, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("UDPClient.New: %s", r.(error).Error())
		}
	}()
	endpoint := host + ":" + strconv.FormatInt(int64(port), 10)
	client = UDPClient{endpoint}
	return
}

// Send a byte array via udp
func (client UDPClient) Send(data []byte) (err error) {
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
	if client.endpoint == "" {
		err = errors.New("UDPClient.Send: endpoint undefined")
		return
	}
	conn, err = net.Dial("udp", client.endpoint)
	if err != nil {
		err = fmt.Errorf("UDPClient.Send: %s", err.Error())
		return
	}
	var sentBytes int
	sentBytes, err = conn.Write(data)
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
