package component

import (
	"bufio"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
)

type Request struct {
	Action string                 `json:"action"`
	Meta   map[string]interface{} `json:"meta"`
}

type Response struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}

type ServerConnector struct {
	conn        net.Conn
	initialized bool
}

// TODO Fetch server information from config
func (this *ServerConnector) Connect() error {
	conn, err := net.Dial("tcp", "localhost:8068")
	if err != nil {
		return err
	}

	this.initialized = true
	this.conn = conn

	return nil
}

func (this *ServerConnector) Disconnect() {
	// Only close connection when the connector was initialized successfully
	if this.initialized {
		this.conn.Close()
	}
}

func (this *ServerConnector) SendRequest(request Request) error {
	if !this.initialized {
		return errors.New("cannot send request, connector must be initialized before use")
	}

	// Hash the token if it is included in the request
	if _, ok := request.Meta["token"]; ok {
		request.Meta["token"] = stringHashB64(request.Meta["token"].(string))
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return err
	}

	this.conn.Write(append([]byte(payload), '\n'))
	return nil
}

func (this *ServerConnector) GetResponse() (Response, error) {
	res, _ := bufio.NewReader(this.conn).ReadString('\n')
	res = strings.TrimSpace(res)

	fmt.Printf("got response: %s\n", res)

	response := Response{}
	if err := json.Unmarshal([]byte(res), &response); err != nil {
		return Response{}, err
	}

	return response, nil
}

/**
 * Shortcut for SendRequest, GetResponse and Disconnect, meant for requests that
 * do not require interactivity with the server.
 */
func (this *ServerConnector) SendAndDisconnect(request Request) (Response, error) {
	defer this.Disconnect()

	err := this.SendRequest(request)
	if err != nil {
		return Response{}, err
	}

	response, err := this.GetResponse()
	if err != nil {
		return Response{}, err
	}

	return response, nil
}

/**
 * Waits until the server sends the given message, or return an error when the
 * received data does not match the message.
 */
func (this *ServerConnector) WaitForMessage(msg string) error {
	res, _ := bufio.NewReader(this.conn).ReadString('\n')
	res = strings.TrimSpace(res)

	fmt.Printf("got message: %s\n", res)

	if res != msg {
		return errors.New(fmt.Sprintf("unexpected message received from server, expected '%s', got '%s'\n", msg, res))
	}

	return nil
}

func stringHashB64(val string) string {
	hash := sha256.New()
	hash.Write([]byte(val))

	return base64.URLEncoding.EncodeToString(hash.Sum(nil))
}
