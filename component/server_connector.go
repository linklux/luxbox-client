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

	"github.com/linklux/luxbox-client/data"
)

type Request struct {
	Action string                 `json:"action"`
	Meta   map[string]interface{} `json:"meta"`
}

type Response struct {
	Code int                    `json:"code"`
	Data map[string]interface{} `json:"data"`
}

type ServerConnector struct {
	conn        net.Conn
	conf        data.Config
	connected   bool
	authEnabled bool
}

func (this *ServerConnector) Connect() error {
	if this.connected {
		return nil
	}

	this.conf = data.GetConfig()
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", this.conf.Server.Host, this.conf.Server.Port))
	if err != nil {
		return errors.New("could not connect to server: " + err.Error())
	}

	this.connected = true
	this.conn = conn

	return nil
}

func (this *ServerConnector) Disconnect() {
	// Only close connection when the connector was initialized successfully
	if this.connected {
		this.conn.Close()
	}
}

func (this *ServerConnector) GetConnection() net.Conn {
	return this.conn
}

/**
 * Control whether or not the user authentication meta data is included in the
 * requests sent to the server with SendRequest(). When enabled, the required
 * meta data is automatically retrieved from the configuration.
 */
func (this *ServerConnector) UserAuthEnabled(enabled bool) {
	this.authEnabled = enabled
}

func (this *ServerConnector) SendRequest(request Request) error {
	if !this.connected {
		return errors.New("cannot send request, not connected to server")
	}

	// Include the authentication meta in the request when enabled
	if this.authEnabled {
		request.Meta["user"] = this.conf.User.User
		request.Meta["token"] = stringHashB64(this.conf.User.Token)
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return err
	}

	this.conn.Write(append([]byte(payload), '\n'))
	return nil
}

func (this *ServerConnector) GetResponse() (Response, error) {
	if !this.connected {
		return Response{}, errors.New("cannot fetch response, not connected to server")
	}

	res, _ := bufio.NewReader(this.conn).ReadString('\n')
	res = strings.TrimSpace(res)

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
