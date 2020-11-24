package nets

import (
	"github.com/jschaf/b2/pkg/errs"
	"net"
	"strconv"
)

type Port = int

// FindAvailablePort returns an available port by asking the kernel for an
// unused port. https://unix.stackexchange.com/a/180500/179300
//
// Copied and slightly modified from https://github.com/phayes/freeport
// Licensed under BSD-3. Copyright (c) 2014, Patrick Hayes
func FindAvailablePort() (p Port, mErr error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer errs.Capturing(&mErr, l.Close, "")
	return l.Addr().(*net.TCPAddr).Port, nil
}

func IsPortOpen(p Port) (r bool, mErr error) {
	l, err := net.Listen("tcp", "localhost:"+strconv.Itoa(p))
	if err != nil {
		return false, nil
	}
	defer errs.Capturing(&mErr, l.Close, "close listener for port open check")
	return true, nil
}
