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
	Control(c CtlCmd) interface{}
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
)

// Pipe ID
const (
	PipeRTSPIN = iota
	PipeEncoder
)

// Pipes global service
var Pipes [3]*PipeSvr

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
	if n, ok := w.Control(CtlCmdName).(string); ok {
		return n
	}

	comm.Error.Fatalf("Worker implements CtlCmdName incorrectly")
	return ""
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
		return w.Encode(pi)
	}

	comm.Error.Fatalf("Worker implements Encoder incorrectly")
	return errBadImplement
}

// SetDecodeSes set Session to Decode
func SetDecodeSes(w Worker, pi *Session) error {
	if w, ok := w.(Decoder); ok {
		return w.Decode(pi)
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
func RPC(url string, cmd string, args interface{}) (map[string]interface{}, error) {

	var message []byte
	var err error
	if message, err = json2.EncodeClientRequest(cmd, args); err != nil {
		comm.Error.Panicf("%v", err)
	}

	var resp *http.Response
	if resp, err = http.Post(url, "application/json", bytes.NewReader(message)); err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	reply := make(map[string]interface{})
	err = json2.DecodeClientResponse(resp.Body, &reply)
	if err != nil {
		return nil, err
	}

	return reply, nil
}

func helperSetMap(m map[string]interface{}, index int, key string, v interface{}) {
	if _, ok := m[key]; ok {
		m[key] = v
	}

	for k := range m {
		if c, ok := m[k].(map[string]interface{}); ok {
			helperSetMap(c, index, key, v)
		} else if c, ok := m[k].([]map[string]interface{}); ok {
			if index < len(c) {
				helperSetMap(c[index], index, key, v)
			}
		}
	}
}
