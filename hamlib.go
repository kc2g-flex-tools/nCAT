package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

type HamlibServer struct {
	sync.RWMutex
	listener net.Listener
	clients  []net.Conn
	handlers map[string]Handler
	exit     chan struct{}
}

type Handler func([]string) string

func NewHamlibServer(listen string) (*HamlibServer, error) {
	l, err := net.Listen("tcp", listen)
	if err != nil {
		return nil, fmt.Errorf("%w while listening on %s", err, listen)
	}

	return &HamlibServer{
		listener: l,
		clients:  []net.Conn{},
		handlers: map[string]Handler{},
		exit:     make(chan struct{}),
	}, nil
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
		conn, err := s.listener.Accept()
		if err != nil {
			goto out
		}
		s.Lock()
		s.clients = append(s.clients, conn)
		s.Unlock()
		go s.handleClient(conn)
	}
out:
	return
}

func (s *HamlibServer) handleClient(conn net.Conn) {
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

func (s *HamlibServer) handleCmd(conn net.Conn, line string) bool {
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

	parts := strings.Split(rest, " ")
	fmt.Printf("%s %v\n", cmd, parts)
	s.RLock()
	defer s.RUnlock()
	if cmd == "q" {
		return true
	}

	ret := "RPRT 1\n" // unknown command
	handler, ok := s.handlers[cmd]
	if ok {
		ret = handler(parts)
	}
	conn.Write([]byte(ret))
	return false
}

func (s *HamlibServer) AddHandler(handler Handler, names ...string) {
	s.Lock()
	defer s.Unlock()
	for _, name := range names {
		s.handlers[name] = handler
	}
}
