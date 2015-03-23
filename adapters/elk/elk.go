package elk

import (
	"errors"
	"io"
	"log"
	"log/syslog"
	"net"
	"os"
	"strings"
	"time"

	"github.com/delectable/logspout/router"
)

func init() {
	router.AdapterFactories.Register(NewElkAdapter, "elk")
}

func getopt(name, dfault string) string {
	value := os.Getenv(name)
	if value == "" {
		value = dfault
	}
	return value
}

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

func (a *ElkAdapter) Stream(logstream chan *router.Message) {
	for message := range logstream {
		io.WriteString(a.conn, message.Data)
		// err := a.tmpl.Execute(a.conn, &ElkMessage{message, a})
		// if err != nil {
		// 	log.Println("syslog:", err)
		// 	a.route.Close()
		// 	return
		// }
	}
}

type ElkMessage struct {
	*router.Message
	adapter *ElkAdapter
}

func (m *ElkMessage) Priority() syslog.Priority {
	switch m.Message.Source {
	case "stdout":
		return syslog.LOG_USER | syslog.LOG_INFO
	case "stderr":
		return syslog.LOG_USER | syslog.LOG_ERR
	default:
		return syslog.LOG_DAEMON | syslog.LOG_INFO
	}
}

func (m *ElkMessage) Hostname() string {
	h, _ := os.Hostname()
	return h
}

func (m *ElkMessage) CleanData() string {
	return strings.Replace(m.Data, "\n", " ", -1)
}

func (m ElkMessage) TestStr() string {
	return "test_string"
}

func (m *ElkMessage) LocalAddr() string {
	return m.adapter.conn.LocalAddr().String()
}

func (m *ElkMessage) Timestamp() string {
	return m.Message.Time.Format(time.RFC3339)
}

func (m *ElkMessage) ContainerName() string {
	return m.Message.Container.Name[1:]
}
