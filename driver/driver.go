// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

// Package driver define driver operation. all sub-cards'
// driver should be in this package
package driver

import "net"

// SlotID is sub-card's Physical SlotID Number
type SlotID int

// WorkerID is sub-card's WorkID, WorkID is uniqual in one sub
// card scope
type WorkerID int

// NameID is sub-card's NameID
type NameID string

// IP is sub-card's IP
type IP net.IP

// IsRunning is Worker's status, on/off
type IsRunning bool

// Resource is shared between path
type Resource string

// CtlCmd ID style const
const (
	CtlCmdStart = iota
	CtlCmdStop
)

// CtlCmd is ID style type for control()
type CtlCmd int

// Card defines sub-cards
type Card interface {
	Open(s SlotID, IP IP) []Worker
	Close() error
}

// Worker defines sub-cards basic operation
type Worker interface {
	Control(c CtlCmd) interface{}
}

// Encoder defines Encoder family operation
type Encoder interface {
	Worker
	Encoder() Resource
}

// String wrappers net.IP.String
func (ip IP) String() string {
	return net.IP(ip).String()
}
