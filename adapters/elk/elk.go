package elk

import (
	"errors"
	// "fmt"
	"io"
	// "log/syslog"
	"net"
	"os"
	// "strings"
	// "time"
	"encoding/json"

	"github.com/delectable/logspout/router"
)

func init() {
	router.AdapterFactories.Register(NewElkAdapter, "elk")
}

// func getopt(name, dfault string) string {
// 	value := os.Getenv(name)
// 	if value == "" {
// 		value = dfault
// 	}
// 	return value
// }

func NewElkAdapter(route *router.Route) (router.LogAdapter, error) {
	transport, found := router.AdapterTransports.Lookup(route.AdapterTransport("udp"))
	if !found {
		return nil, errors.New("unable to find adapter: " + route.Adapter)
	}
	conn, err := transport.Dial(route.Address, route.Options)
	if err != nil {
		return nil, err
	}

	return &ElkAdapter{
		route: route,
		conn:  conn,
	}, nil
}

type ElkAdapter struct {
	conn  net.Conn
	route *router.Route
}

func (adapter *ElkAdapter) Stream(logstream chan *router.Message) {
	for message := range logstream {
		elkMessage := NewElkMessage(message)
		io.WriteString(adapter.conn, elkMessage.ToString())
	}
}

type ElkMessage struct {
	routerMessage *router.Message
	Object        struct {
		Time     int64  `json: "time"`
		Message  string `json: "message"`
		Hostname string `json: "hostname"`
	}
}

func NewElkMessage(routerMessage *router.Message) *ElkMessage {
	elkMessage := &ElkMessage{
		routerMessage: routerMessage,
	}

	elkMessage.Object.Time = routerMessage.Time.Unix()
	elkMessage.Object.Message = routerMessage.Data
	elkMessage.Object.Hostname, _ = os.Hostname()

	return elkMessage
}

func (elkMessage *ElkMessage) ToString() string {
	return_string, _ := json.Marshal(elkMessage.Object)
	return string(return_string)
}
