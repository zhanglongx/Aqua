// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

// Package driver define driver operation. all sub-cards'
// driver should be in this package
package driver

import (
	"errors"
	"net"

	"github.com/zhanglongx/Aqua/comm"
)

// CtlCmd ID style const
const (
	CtlCmdStart = iota
	CtlCmdStop
	CtlCmdName
	CtlCmdIP
	CtlCmdSlot
	CtlCmdWorkerID
)

// CtlCmd is ID style type for control()
type CtlCmd int

// Card defines sub-cards
type Card interface {
	Open(s int, IP net.IP) ([]Worker, error)
	Close() error
}

// Worker defines generic operation
type Worker interface {
	Control(c CtlCmd) interface{}
}

// Encoder defines Encoder family operation
type Encoder interface {
	Worker
	Encoder(*Session) error
}

// Decoder defines Decoder family operation
type Decoder interface {
	Worker
	Decoder(*Session) error
}

var (
	errBadImplement = errors.New("Bad Implement")
)

// GetWorkerName get Worker's Name
func GetWorkerName(w Worker) string {
	if n, ok := w.Control(CtlCmdName).(string); ok {
		return n
	}

	comm.Error.Fatalf("Worker implements CtlCmdName incorrectly")
	return ""
}

// GetWorkerSlot get Worker's Slot
func GetWorkerSlot(w Worker) int {
	if s, ok := w.Control(CtlCmdSlot).(int); ok {
		return s
	}

	comm.Error.Fatalf("Worker implements CtlCmdSlot incorrectly")
	return 0
}

// GetWorkerWorkerID get Worker's Slot
func GetWorkerWorkerID(w Worker) int {
	if s, ok := w.Control(CtlCmdWorkerID).(int); ok {
		return s
	}

	comm.Error.Fatalf("Worker implements CtlCmdWorkerID incorrectly")
	return 0
}

// GetWorkerWorkerIP get Worker's Slot
func GetWorkerWorkerIP(w Worker) net.IP {
	if IP, ok := w.Control(CtlCmdIP).(net.IP); ok {
		return IP
	}

	comm.Error.Fatalf("Worker implements CtlCmdWorkerIP incorrectly")
	return net.IPv4(0, 0, 0, 0)
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

	return nil
}

// SetEncodeSes set Session to Encoder
func SetEncodeSes(w Worker, pi *Session) error {
	if w, ok := w.(Encoder); ok {
		return w.Encoder(pi)
	}

	comm.Error.Fatalf("Worker implements Encoder incorrectly")
	return errBadImplement
}

// SetDecodeSes set Session to Decode
func SetDecodeSes(w Worker, pi *Session) error {
	if w, ok := w.(Decoder); ok {
		return w.Decoder(pi)
	}

	comm.Error.Fatalf("Worker implements Decoder incorrectly")
	return errBadImplement
}

// IsWorkerDec return bool
func IsWorkerDec(w Worker) bool {
	if _, ok := w.(Decoder); ok {
		return true
	}

	return false
}

// IsWorkerEnc return bool
func IsWorkerEnc(w Worker) bool {
	if _, ok := w.(Encoder); ok {
		return true
	}

	return false
}
