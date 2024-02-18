package net

import (
	"fmt"
	"net"
	"sync"
)

type ListenAddr struct {
	mu sync.Mutex

	address  string
	listener net.Listener
}

func (a *ListenAddr) String() string {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a == nil {
		return fmt.Sprintf("<nil>")
	}

	if a.listener == nil {
		return fmt.Sprintf("provided but not bound: %v", a.address)
	}

	return a.listener.Addr().String()
}

func (a *ListenAddr) Listener() (net.Listener, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

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

func (a *ListenAddr) UnmarshalYAML(unmarshal func(interface{}) error) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	var address string
	if err := unmarshal(&address); err != nil {
		return err
	}

	a.address = address

	return nil
}

func ListenAddrFromAddress(addr string) *ListenAddr {
	return &ListenAddr{
		address: addr,
	}
}
