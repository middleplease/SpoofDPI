package packet

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"strings"
)

var validMethod = map[string]struct{}{
	"DELETE":      {},
	"GET":         {},
	"HEAD":        {},
	"POST":        {},
	"PUT":         {},
	"CONNECT":     {},
	"OPTIONS":     {},
	"TRACE":       {},
	"COPY":        {},
	"LOCK":        {},
	"MKCOL":       {},
	"MOVE":        {},
	"PROPFIND":    {},
	"PROPPATCH":   {},
	"SEARCH":      {},
	"UNLOCK":      {},
	"BIND":        {},
	"REBIND":      {},
	"UNBIND":      {},
	"ACL":         {},
	"REPORT":      {},
	"MKACTIVITY":  {},
	"CHECKOUT":    {},
	"MERGE":       {},
	"M-SEARCH":    {},
	"NOTIFY":      {},
	"SUBSCRIBE":   {},
	"UNSUBSCRIBE": {},
	"PATCH":       {},
	"PURGE":       {},
	"MKCALENDAR":  {},
	"LINK":        {},
	"UNLINK":      {},
}

type HttpPacket struct {
	raw     []byte
	method  string
	domain  string
	port    string
	path    string
	version string
}

func ReadHttpPacket(rdr io.Reader) (*HttpPacket, error) {
	p, err := parse(rdr)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (p *HttpPacket) Raw() []byte {
	return p.raw
}
func (p *HttpPacket) Method() string {
	return p.method
}

func (p *HttpPacket) Domain() string {
	return p.domain
}

func (p *HttpPacket) Port() string {
	return p.port
}

func (p *HttpPacket) Version() string {
	return p.version
}

func (p *HttpPacket) IsValidMethod() bool {
	if _, exists := validMethod[p.Method()]; exists {
		return true
	}

	return false
}

func (p *HttpPacket) IsConnectMethod() bool {
	return p.Method() == "CONNECT"
}

func (p *HttpPacket) Tidy() {
	s := string(p.raw)

	lines := strings.Split(s, "\r\n")

	lines[0] = p.method + " " + p.path + " " + p.version

	for i := 0; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "Proxy-Connection") {
			lines[i] = ""
		}
	}

	result := ""

	for i := 0; i < len(lines); i++ {
		if lines[i] == "" {
			continue
		}

		result += lines[i] + "\r\n"
	}

	result += "\r\n"

	p.raw = []byte(result)
}

func parse(rdr io.Reader) (*HttpPacket, error) {
	sb := strings.Builder{}
	tee := io.TeeReader(rdr, &sb)
	request, err := http.ReadRequest(bufio.NewReader(tee))
	if err != nil {
		return nil, err
	}

	p := &HttpPacket{}
	p.raw = []byte(sb.String())

	p.domain, p.port, err = net.SplitHostPort(request.Host)
	if err != nil {
		p.domain = request.Host
		p.port = ""
	}

	p.method = request.Method
	p.version = request.Proto
	p.path = request.URL.Path

	if request.URL.RawQuery != "" {
		p.path += "?" + request.URL.RawQuery
	}

	if request.URL.RawFragment != "" {
		p.path += "#" + request.URL.RawFragment
	}
	if p.path == "" {
		p.path = "/"
	}

	request.Body.Close()
	return p, nil
}
