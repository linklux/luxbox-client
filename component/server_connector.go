package component

import (
	"bufio"
	"encoding/json"
	"net"
)

type Request struct {
	Action string            `json:"action"`
	Meta   map[string]string `json:"meta"`
}

type Response struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}

type ServerConnector struct {
}

func (this *ServerConnector) Send(request Request) (Response, error) {
	conn, err := net.Dial("tcp", "localhost:8068")
	if err != nil {
		return Response{}, err
	}

	defer conn.Close()

	payload, err := json.Marshal(request)
	if err != nil {
		return Response{}, err
	}

	// send to server
	conn.Write(append([]byte(payload), '\n'))

	// wait for reply
	res, _ := bufio.NewReader(conn).ReadString('\n')
	response := Response{}

	if err = json.Unmarshal([]byte(res), &response); err != nil {
		return Response{}, err
	}

	return response, nil
}
