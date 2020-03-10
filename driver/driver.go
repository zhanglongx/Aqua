// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

// Package driver define driver operation. all sub-cards'
// driver should be in this package
package driver

import (
	"net"

	"github.com/zhanglongx/Aqua/comm"
)

// CtlCmd ID style const
const (
	CtlCmdStart = iota
	CtlCmdStop
	CtlCmdName
)

// CtlCmd is ID style type for control()
type CtlCmd int

// Card defines sub-cards
type Card interface {
	Open(s int, IP net.IP) []Worker
	Close() error
}

// Worker defines generic operation
type Worker interface {
	Control(c CtlCmd) interface{}
}

// Encoder defines Encoder family operation
type Encoder interface {
	Worker
	Encoder() Resource
}

// GetWorkerName return Worker's Name
func GetWorkerName(w Worker) string {
	if n, ok := w.Control(CtlCmdName).(string); ok {
		return n
	}

	comm.Error.Printf("worker implements CtlCmdName incorrectly")
	return ""
}

// SetWorkerRunning set Running status
func SetWorkerRunning(w Worker, r bool) error {
	if r {
		if err, ok := w.Control(CtlCmdStart).(error); ok {
			return err
		}
	} else {
		if err, ok := w.Control(CtlCmdStop).(error); ok {
			return err
		}
	}

	comm.Error.Printf("worker implements CtlCmdStart/Stop incorrectly")
	return nil
}
