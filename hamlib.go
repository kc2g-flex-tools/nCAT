package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
)

type HamlibServer struct {
	sync.RWMutex
	listener net.Listener
	clients  []net.Conn
	handlers map[string]interface{}
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

type HandlerFunc func(context.Context, []string) (string, error)

type Handler struct {
	Name         string
	ShortName    string
	cb           HandlerFunc
	requiredArgs []string
	minArgs      *int
	maxArgs      *int
	allArgs      bool
	errResponse  *string
	fieldNames   []string
}

type Option interface {
	apply(h *Handler)
}

type requiredArgs []string

func (ra requiredArgs) apply(h *Handler) {
	x := []string(ra)
	h.requiredArgs = x
}
func RequiredArgs(ra ...string) requiredArgs {
	return requiredArgs(ra)
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

type fieldNames []string

func (fn fieldNames) apply(h *Handler) {
	x := []string(fn)
	h.fieldNames = x
}
func FieldNames(fn ...string) fieldNames {
	return fieldNames(fn)
}

func NewHandler(name, shortName string, cb HandlerFunc, opts ...Option) Handler {
	h := Handler{
		Name:      name,
		ShortName: shortName,
		cb:        cb,
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
	}
}

var Success = "RPRT 0\n"
var Error = "RPRT 1\n"

type Conn struct {
	net.Conn
	// New enough hamlib will send \chk_vfo before \dump_state, keep track of whether it has.
	chkVFOexecuted bool
}

func NewConn(netConn net.Conn) Conn {
	return Conn{
		Conn: netConn,
	}
}

func (s *HamlibServer) Listen(listen string) error {
	l, err := net.Listen("tcp", listen)
	if err != nil {
		return fmt.Errorf("%w while listening on %s", err, listen)
	}

	s.listener = l
	return nil
}

func (s *HamlibServer) Run(ctx context.Context) {
	go func() {
		<-ctx.Done()
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
		conn := NewConn(netConn)

		s.Lock()
		s.clients = append(s.clients, conn)
		s.Unlock()
		clientCtx := context.WithValue(ctx, connKey{}, &conn)
		go s.handleClient(clientCtx, &conn)
	}
out:
	return
}

func (s *HamlibServer) handleClient(ctx context.Context, conn *Conn) {
	lines := bufio.NewScanner(conn)
	for lines.Scan() {
		exit := s.handleCmd(ctx, conn, lines.Text())
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
	conn.Close()
	s.Unlock()
}

func (s *HamlibServer) handleCmd(ctx context.Context, conn *Conn, line string) bool {
	if line == "" {
		return false
	}

	cmd := line[0:1]
	var rest string

	if cmd == "\\" {
		line = line[1:]
		spaceIdx := strings.Index(line, " ")
		if spaceIdx == -1 {
			cmd = line
			rest = ""
		} else {
			cmd = line[:spaceIdx]
			rest = line[spaceIdx+1:]
		}
	} else {
		cmd = "short:" + cmd
		rest = strings.TrimLeft(line[1:], " ")
	}

	parts := []string{cmd}
	if rest != "" {
		parts = append(parts, strings.Split(rest, " ")...)
	}
	log.Trace().Strs("cmd", parts).Msg("")
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
				handlerCtx := context.WithValue(ctx, handlerKey{}, &handler)
				ret, e = handler.cb(handlerCtx, args)
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
				log.Warn().Strs("cmd", parts).Err(e).Msg("Handler returned error")
			}
		case nil:
			log.Warn().Strs("cmd", parts[:i+1]).Msg("No handler found")
		default:
			log.Warn().Strs("cmd", parts[:i+1]).Interface("handler", handler).Msg("Found an unknown thing in the handler table")
		}
		break
	}
	log.Trace().Str("response", ret).Send()
	conn.Write([]byte(ret))
	return false
}

func (s *HamlibServer) AddHandler(handler Handler) {
	s.Lock()
	defer s.Unlock()

	names := []string{handler.Name}
	if handler.ShortName != "" {
		names = append(names, "short:"+handler.ShortName)
	}

	for _, name := range names {
		table := s.handlers
		nameWithArgs := []string{name}
		if len(handler.requiredArgs) > 0 {
			nameWithArgs = append(nameWithArgs, handler.requiredArgs...)
		}

		for _, part := range nameWithArgs[:len(nameWithArgs)-1] {
			if table[part] == nil {
				table[part] = map[string]interface{}{}
			}
			table = table[part].(map[string]interface{})
		}

		table[nameWithArgs[len(nameWithArgs)-1]] = handler
	}
}
