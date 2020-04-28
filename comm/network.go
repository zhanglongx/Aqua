// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

package comm

import (
	"net"
	"runtime"
	"strings"
)

// NetCfg mainly wrappers Ifconfig
type NetCfg struct {
	Name string

	ip net.IP
}

// create reads Interfaces
func (n *NetCfg) create() error {

	i, err := net.InterfaceByName(n.Name)
	if err != nil {
		return err
	}

	addrs, _ := i.Addrs()
	for _, a := range addrs {
		// Must drop the stuff after the slash in order to convert it to an IP instance
		split := strings.Split(a.String(), "/")
		addrStr0 := split[0]

		// Parse the string to an IP instance
		ip := net.ParseIP(addrStr0)
		if ip.To4() == nil {
			continue
		}

		n.ip = ip
	}

	if n.ip.String() == net.IPv4zero.String() {
		panic("Interface error")
	}

	return nil
}

// GetIPv4 return a net.IP assigned to hw
func (n *NetCfg) GetIPv4() net.IP {
	return n.ip
}

// SetIPv4 set hw to IP
func (n *NetCfg) SetIPv4(ip net.IP) error {
	if runtime.GOOS == "windows" {

	}

	// TODO:

	return nil
}
