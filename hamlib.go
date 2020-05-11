package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

type HamlibServer struct {
	sync.RWMutex
	listener net.Listener
	clients  []net.Conn
	handlers map[string]interface{}
	exit     chan struct{}
}

type customError struct {
	error
	response string
}

func CustomError(err error, response string) error {
	return customError{
		error:    err,
		response: response,
	}
}

type HandlerFunc func(Conn, []string) (string, error)

type Handler struct {
	cb          HandlerFunc
	minArgs     *int
	maxArgs     *int
	allArgs     bool
	errResponse *string
}

type Option interface {
	apply(h *Handler)
}

type MinArgs int

func (ma MinArgs) apply(h *Handler) {
	x := int(ma)
	h.minArgs = &x
}

type MaxArgs int

func (ma MaxArgs) apply(h *Handler) {
	x := int(ma)
	h.maxArgs = &x
}

type Args int

func (a Args) apply(h *Handler) {
	x := int(a)
	h.minArgs = &x
	h.maxArgs = &x
}

type AllArgs bool

func (aa AllArgs) apply(h *Handler) {
	h.allArgs = bool(aa)
}

type ErrResponse string

func (er ErrResponse) apply(h *Handler) {
	x := string(er)
	h.errResponse = &x
}

type names [][]string

func NewHandler(cb HandlerFunc, opts ...Option) Handler {
	h := Handler{
		cb: cb,
	}

	for _, o := range opts {
		o.apply(&h)
	}

	return h
}

func NewHamlibServer() *HamlibServer {
	return &HamlibServer{
		clients:  []net.Conn{},
		handlers: map[string]interface{}{},
		exit:     make(chan struct{}),
	}
}

var Success = "RPRT 0\n"
var Error = "RPRT 1\n"

type Conn struct {
	net.Conn
}

func (s *HamlibServer) Listen(listen string) error {
	l, err := net.Listen("tcp", listen)
	if err != nil {
		return fmt.Errorf("%w while listening on %s", err, listen)
	}

	s.listener = l
	return nil
}

func (s *HamlibServer) Close() {
	close(s.exit)
}

func (s *HamlibServer) Run() {
	go func() {
		<-s.exit
		s.RLock()
		defer s.RUnlock()

		s.listener.Close()
		for _, client := range s.clients {
			client.Close()
		}
	}()

	for {
		netConn, err := s.listener.Accept()
		if err != nil {
			goto out
		}
		conn := Conn{
			Conn: netConn,
		}

		s.Lock()
		s.clients = append(s.clients, conn)
		s.Unlock()
		go s.handleClient(conn)
	}
out:
	return
}

func (s *HamlibServer) handleClient(conn Conn) {
	lines := bufio.NewScanner(conn)
	for lines.Scan() {
		exit := s.handleCmd(conn, lines.Text())
		if exit {
			break
		}
	}
	s.Lock()
	for i, cl := range s.clients {
		if cl == conn {
			s.clients = append(s.clients[:i], s.clients[i+1:]...)
		}
	}
	s.Unlock()
}

func (s *HamlibServer) handleCmd(conn Conn, line string) bool {
	if line == "" {
		return false
	}

	cmd := line[0:1]
	rest := strings.TrimLeft(line[1:], " ")

	if cmd == "\\" {
		spaceIdx := strings.Index(line, " ")
		if spaceIdx == -1 {
			cmd = line
			rest = ""
		} else {
			cmd = line[:spaceIdx]
			rest = line[spaceIdx+1:]
		}
	}

	parts := []string{cmd}
	if rest != "" {
		parts = append(parts, strings.Split(rest, " ")...)
	}
	log.Println(parts)
	s.RLock()
	defer s.RUnlock()
	if cmd == "q" {
		return true
	}

	ret := Error // unknown command
	table := s.handlers
	i := 0

	for {
		h := table[parts[i]]
		switch handler := h.(type) {
		case map[string]interface{}:
			table = handler
			i += 1
			continue
		case Handler:
			var e error
			var args []string
			if handler.allArgs {
				args = parts
			} else {
				args = parts[i+1:]
			}

			if handler.minArgs != nil && len(args) < *handler.minArgs {
				e = fmt.Errorf("required at least %d args, got %d", *handler.minArgs, len(args))
			} else if handler.maxArgs != nil && len(args) > *handler.maxArgs {
				e = fmt.Errorf("required max %d args, got %d", *handler.maxArgs, len(args))
			} else {
				ret, e = handler.cb(conn, args)
			}

			if e != nil {
				switch err := e.(type) {
				case customError:
					ret = err.response
				default:
					if handler.errResponse != nil {
						ret = *handler.errResponse
					} else {
						ret = Error
					}
				}
				log.Println("Handler returned error:", e)
			}
		case nil:
			log.Println("No handler found for command", parts[:i+1])
		default:
			log.Printf("Found an unknown thing in the handler table: %T\n", handler)
		}
		break
	}
	conn.Write([]byte(ret))
	return false
}

func (s *HamlibServer) AddHandler(names names, handler Handler) {
	s.Lock()
	defer s.Unlock()

	for _, name := range names {
		table := s.handlers
		for _, part := range name[:len(name)-1] {
			if table[part] == nil {
				table[part] = map[string]interface{}{}
			}
			table = table[part].(map[string]interface{})
		}

		table[name[len(name)-1]] = handler
	}
}
