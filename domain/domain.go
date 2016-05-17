package domain

import (
	"net"
	"time"
)

type Port struct {
	Source      net.IP
	Destination net.IP
}

type Status int

const (
	ALL Status = 1 + iota
	RUNNING
	STOPPED
	UNKNOWN
)

type Container struct {
	Id          string
	Image       string
	Created     time.Time
	Datestarted time.Time
	Ports       []Port
	Status      Status
	Name        string
	Labels      []string
	IP          net.IP
}

type Client interface {
}

type Parser interface {
}
