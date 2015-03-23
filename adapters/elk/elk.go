package elk

import (
	"errors"
	"fmt"
	"io"
	// "log/syslog"
	"net"
	"os"
	// "strings"
	// "time"

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

func (adapter *ElkAdapter) Stream(logstream chan *router.Message) {
	for message := range logstream {
		output_string := fmt.Sprintf("{entry: {time: %d, message: %s}}", message.Time.Unix, message.Data)
		io.WriteString(adapter.conn, output_string)
		// err := a.tmpl.Execute(a.conn, &ElkMessage{message, a})
		// if err != nil {
		// 	log.Println("syslog:", err)
		// 	a.route.Close()
		// 	return
		// }
	}
}

// func (m *ElkMessage) Priority() syslog.Priority {
// 	switch m.Message.Source {
// 	case "stdout":
// 		return syslog.LOG_USER | syslog.LOG_INFO
// 	case "stderr":
// 		return syslog.LOG_USER | syslog.LOG_ERR
// 	default:
// 		return syslog.LOG_DAEMON | syslog.LOG_INFO
// 	}
// }

// func (m *ElkMessage) ContainerName() string {
// 	return m.Message.Container.Name[1:]
// }
