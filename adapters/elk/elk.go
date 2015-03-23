package elk

import (
	"errors"
	"fmt"
	"log"
	"log/syslog"
	"net"
	"os"
	"text/template"
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

	// priority := getopt("ELK_PRIORITY", "{{.Priority}}")
	// hostname := getopt("ELK_HOSTNAME", "{{.Container.Config.Hostname}}")
	// pid := getopt("ELK_PID", "{{.Container.State.Pid}}")
	// tag := getopt("ELK_TAG", "{{.ContainerName}}"+route.Options["append_tag"])
	// structuredData := getopt("ELK_STRUCTURED_DATA", "")
	// if route.Options["structured_data"] != "" {
	// 	structuredData = route.Options["structured_data"]
	// }
	data := getopt("ELK_DATA", "{{.Data}}")

	// tmplStr := fmt.Sprintf("CRUNCHY BACON: <%d> {{.Timestamp}} %s %s %d - [%s] %s",
	// 	priority, hostname, tag, pid, structuredData, data)

	// fmt.Println("GOT A LOG ENTRY.")

	tmplStr := fmt.Sprintf("CRUNCHY BACON: {{.Timestamp}} %s END", data)

	tmpl, err := template.New("elk").Parse(tmplStr)
	if err != nil {
		return nil, err
	}
	return &ElkAdapter{
		route: route,
		conn:  conn,
		tmpl:  tmpl,
	}, nil
}

type ElkAdapter struct {
	conn  net.Conn
	route *router.Route
	tmpl  *template.Template
}

func (a *ElkAdapter) Stream(logstream chan *router.Message) {
	for message := range logstream {
		err := a.tmpl.Execute(a.conn, &ElkMessage{message, a})
		if err != nil {
			log.Println("syslog:", err)
			a.route.Close()
			return
		}
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

func (m *ElkMessage) LocalAddr() string {
	return m.adapter.conn.LocalAddr().String()
}

func (m *ElkMessage) Timestamp() string {
	return m.Message.Time.Format(time.RFC3339)
}

func (m *ElkMessage) ContainerName() string {
	return m.Message.Container.Name[1:]
}
