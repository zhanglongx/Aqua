// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

// Package driver define driver operation. all sub-cards'
// driver should be in this package
package driver

import (
	"bytes"
	"errors"
	"net"
	"net/http"

	"github.com/gorilla/rpc/v2/json2"
	"github.com/zhanglongx/Aqua/comm"
)

// CtlCmd ID style const
const (
	CtlCmdStart = iota
	CtlCmdStop
	CtlCmdName
	CtlCmdIP
	CtlCmdWorkerID
	CtlCmdSetting
)

// CtlCmd is ID style type for control()
type CtlCmd int

// Card defines sub-cards
type Card interface {
	Open() ([]Worker, error)
	Close() error
}

// Worker defines generic operation
type Worker interface {
	Control(c CtlCmd, arg interface{}) interface{}
}

// Encoder defines Encoder family operation
type Encoder interface {
	Worker
	Encode(sess *Session) error
}

// Decoder defines Decoder family operation
type Decoder interface {
	Worker
	Decode(sess *Session) error
}

var (
	errBadImplement = errors.New("Bad Implement")
	errTypeError    = errors.New("Type Error")
	errKeyError     = errors.New("Key Error")
)

// Pipe ID
const (
	PipeRTSPIN = iota
	PipeEncoder
)

// Pipes global service
var Pipes [2]*PipeSvr

var transitSvr transit

// Init create PipeSvr
func init() {

	CfgTransit := comm.AppCfg.TransitSvr

	transitSvr = transit{selfIP: CfgTransit}

	Pipes[PipeRTSPIN] = &PipeSvr{IP: CfgTransit, Prefix: 0}
	Pipes[PipeEncoder] = &PipeSvr{IP: CfgTransit, Prefix: 1000}

	Pipes[PipeRTSPIN].Create()
	Pipes[PipeEncoder].Create()
}

// GetWorkerName get Worker's Name
func GetWorkerName(w Worker) string {
	if n, ok := w.Control(CtlCmdName, nil).(string); ok {
		return n
	}

	comm.Error.Fatalf("Worker implements CtlCmdName incorrectly")
	return ""
}

// GetWorkerWorkerID get Worker's Slot
func GetWorkerWorkerID(w Worker) int {
	if s, ok := w.Control(CtlCmdWorkerID, nil).(int); ok {
		return s
	}

	comm.Error.Fatalf("Worker implements CtlCmdWorkerID incorrectly")
	return 0
}

// GetWorkerWorkerIP get Worker's Slot
func GetWorkerWorkerIP(w Worker) net.IP {
	if IP, ok := w.Control(CtlCmdIP, nil).(net.IP); ok {
		return IP
	}

	comm.Error.Fatalf("Worker implements CtlCmdWorkerIP incorrectly")
	return net.IPv4(0, 0, 0, 0)
}

// SetWorkerRunning set Running status
func SetWorkerRunning(w Worker, r bool) error {
	if r {
		if err, ok := w.Control(CtlCmdStart, nil).(error); ok {
			return err
		}
	} else {
		if err, ok := w.Control(CtlCmdStop, nil).(error); ok {
			return err
		}
	}

	return nil
}

// SetWorkerSettings set Running status
func SetWorkerSettings(w Worker, s map[string]interface{}) error {
	if err, ok := w.Control(CtlCmdSetting, s).(error); ok {
		return err
	}

	return nil
}

// SetEncodeSes set Session to Encoder
func SetEncodeSes(w Worker, sess *Session) error {
	if w, ok := w.(Encoder); ok {
		return w.Encode(sess)
	}

	comm.Error.Fatalf("Worker implements Encoder incorrectly")
	return errBadImplement
}

// SetDecodeSes set Session to Decode
func SetDecodeSes(w Worker, sess *Session) error {
	if w, ok := w.(Decoder); ok {
		return w.Decode(sess)
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

// RPC wrappers JSON-rpc queries
func RPC(url string, cmd string, args interface{}, reply interface{}) error {

	var message []byte
	var err error
	if message, err = json2.EncodeClientRequest(cmd, args); err != nil {
		comm.Error.Panicf("%v", err)
	}

	var resp *http.Response
	if resp, err = http.Post(url, "application/json", bytes.NewReader(message)); err != nil {
		return err
	}

	defer resp.Body.Close()

	if err = json2.DecodeClientResponse(resp.Body, reply); err != nil {
		return err
	}

	return nil
}

// helperSetMap lookup key in m, and change the value. If value is a slice, index
// will be used. All keys with same name in sub-level will be changes.
// TODO: return err if key not exist
func helperSetMap(m map[string]interface{}, index int, key string, v interface{}) {
	if _, ok := m[key]; ok {
		m[key] = v
	}

	for k := range m {
		if c, ok := m[k].(map[string]interface{}); ok {
			helperSetMap(c, index, key, v)
		} else if c, ok := m[k].([]interface{}); ok {
			if index < len(c) {
				if cc, ok := c[index].(map[string]interface{}); ok {
					helperSetMap(cc, index, key, v)
				}
			}
		}
	}
}
