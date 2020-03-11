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
	Encoder() (InnerRes, error)
}

// Decoder defines Decoder family operation
type Decoder interface {
	Worker
	Decoder(InnerRes) error
}

var (
	errBadImplement = errors.New("Bad Implement")
)

// GetWorkerName get Worker's Name
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

	return nil
}

// GetEncodeRes get Encoder's Res
func GetEncodeRes(w Worker) (InnerRes, error) {
	if w, ok := w.(Encoder); ok {
		return w.Encoder()
	}

	comm.Error.Printf("worker implements Encoder incorrectly")
	return InnerRes{}, errBadImplement
}

// SetDecodeRes set Res to Decode
func SetDecodeRes(w Worker, ir InnerRes) error {
	if w, ok := w.(Decoder); ok {
		return w.Decoder(ir)
	}

	comm.Error.Printf("worker implements Decoder incorrectly")
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

// helperTrsInPort
func helperTrsInPort(s int, w int) int {
	if (s < 0 || s > 16) || (w < 0 || w > 16) {
		comm.Error.Fatal("slot port error")
		return -1
	}

	return 8000 + 64*s + w
}
