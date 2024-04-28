package main

import (
	"testing"
	"time"
)

var (
	transferServer = "10.190.50.53:6060"
	graphServer    = "10.190.50.53:6071"
	timeout        = 5 * time.Second
)

func TestNewFalconClient(t *testing.T) {
	_, err := NewFalconClient(transferServer, timeout)
	if err != nil {
		t.Error(err)
	}
}

func TestHealth(t *testing.T) {
	c, _ := NewFalconClient(transferServer, timeout)
	_, err := c.Health()
	if err != nil {
		t.Error(err)
	}
}

func TestVersion(t *testing.T) {
	c, _ := NewFalconClient(transferServer, timeout)
	_, err := c.Version()
	if err != nil {
		t.Error(err)
	}
}

func TestTransferCounter(t *testing.T) {
	c, _ := NewFalconClient(transferServer, timeout)
	tc := &FalconTransferClient{*c}
	_, err := tc.Counter()
	if err != nil {
		t.Error(err)
	}
}

func TestGraphHealth(t *testing.T) {
	c, _ := NewFalconClient(graphServer, timeout)
	gc := &FalconGraphClient{*c}
	_, err := gc.Health()
	if err != nil {
		t.Error(err)
	}
}

func TestGraphVersion(t *testing.T) {
	c, _ := NewFalconClient(graphServer, timeout)
	gc := &FalconGraphClient{*c}
	_, err := gc.Version()
	if err != nil {
		t.Error(err)
	}
}

func TestGraphCounter(t *testing.T) {
	c, _ := NewFalconClient(graphServer, timeout)
	gc := &FalconGraphClient{*c}
	_, err := gc.Counter()
	if err != nil {
		t.Error(err)
	}
}

func TestClientFactory(t *testing.T) {
	c, err := ClientFactory("alarm", "10.190.50.53:9912", 5*time.Second)
	if err != nil {
		t.Error(err)
	}

	_, err = c.Version()
	if err != nil {
		t.Error(err)
	}
}
