package measurement

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

type tcpAvailability struct {
	name          string
	summaryPhrase string
	url           string
	port          int
}

func (t *tcpAvailability) Name() string {
	return t.name
}

func (t *tcpAvailability) SummaryPhrase() string {
	return t.summaryPhrase
}

func (t *tcpAvailability) PerformMeasurement() (string, string, string, bool) {
	addr := net.JoinHostPort(t.url, fmt.Sprintf("%d", t.port))
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return err.Error(), "", "", false
	}
	defer conn.Close() //nolint:errcheck

	_, err = conn.Write([]byte("knock-knock"))
	if err != nil {
		return err.Error(), "", "", false
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return err.Error(), "", "", false
	}

	body := string(buf[:n])

	if strings.Contains(body, "Hello from Uptimer.") {
		return "", "", "", true
	}

	return "TCP App not returning expected response", "", "", false
}
