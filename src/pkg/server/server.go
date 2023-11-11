package server

import (
	"fmt"
	"net"
)

type ListenAddr struct {
	address  string
	listener net.Listener
}

func (a *ListenAddr) String() string {
	if a.listener == nil {
		return fmt.Sprintf("provided but not bound: %v", a.address)
	}

	return a.listener.Addr().String()
}

func (a *ListenAddr) Listener() (net.Listener, error) {
	if a.listener != nil {
		return a.listener, nil
	}

	listener, err := net.Listen("tcp", a.address)
	if err != nil {
		return nil, err
	}
	a.listener = listener

	return listener, nil
}

func ListenAddrFromListener(listener net.Listener) *ListenAddr {
	return &ListenAddr{
		address:  listener.Addr().String(),
		listener: listener,
	}
}

func ListenAddrFromAddress(addr string) *ListenAddr {
	return &ListenAddr{
		address: addr,
	}
}
