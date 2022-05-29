package main

import (
	"errors"
	"fmt"
	"net"
	"net/url"
)

type ListenAddress struct {
	network string
	address string
}

func NewListenAddress(value string) (*ListenAddress, error) {
	l := &ListenAddress{}
	err := l.Set(value)
	return l, err
}

func (l *ListenAddress) String() string {
	switch l.network {
	case "tcp", "unix":
		return fmt.Sprintf("%s://%s", l.network, l.address)
	default:
		return ""
	}
}

func (l *ListenAddress) Set(value string) error {
	if value == "" {
		l.network = ""
		l.address = ""
		return nil
	}

	u, err := url.Parse(value)
	if err != nil {
		return err
	}

	switch u.Scheme {
	case "tcp":
		l.network = "tcp"
		l.address = u.Host
	case "unix":
		l.network = "unix"
		l.address = u.Path
	default:
		return errors.New("unacceptable scheme, only 'tcp' and 'unix' are permitted")
	}
	return nil
}

func (l *ListenAddress) Listen() (net.Listener, error) {
	if l.network == "" {
		return nil, errors.New("listen address is not set")
	}
	return net.Listen(l.network, l.address)
}
