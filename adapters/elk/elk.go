package elk

import (
	"errors"
	// "fmt"
	"io"
	"io/ioutil"
	// "log/syslog"
	"net"
	// "os"
	"strings"
	// "time"
	"encoding/json"

	"github.com/delectable/logspout/router"
)

var HOSTNAME string

func init() {
	router.AdapterFactories.Register(NewElkAdapter, "elk")

	hostname_bytestring, _ := ioutil.ReadFile("/etc/hostname") // this should stick around in memory, not run for every log line
	HOSTNAME = strings.TrimSpace(string(hostname_bytestring))

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
		Image    string `json: "image"`
		App      string `json: "app"`
	}
}

func NewElkMessage(routerMessage *router.Message) *ElkMessage {
	elkMessage := &ElkMessage{
		routerMessage: routerMessage,
	}

	elkMessage.Object.Time = routerMessage.Time.Unix()
	elkMessage.Object.Message = routerMessage.Data

	elkMessage.Object.Hostname = HOSTNAME

	elkMessage.Object.Image = routerMessage.Container.Config.Image

	env_map := make(map[string]string)
	for _, blob := range routerMessage.Container.Config.Env {
		split_blob := strings.Split(blob, "=")
		env_map[split_blob[0]] = split_blob[1]
	}

	elkMessage.Object.App = env_map["MARATHON_APP_ID"]

	return elkMessage
}

func (elkMessage *ElkMessage) ToString() string {
	return_string, _ := json.Marshal(elkMessage.Object)
	return string(return_string)
}
