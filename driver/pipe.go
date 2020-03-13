// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

package driver

import (
	"errors"
	"net"
)

// Node alloc Pipe
type Node struct {
	// IP is the Svr IP
	IP net.IP

	// Prefix to identity services
	Prefix int

	all map[int]*pipe
}

// pipe is pipeline shared between workers
type pipe struct {
	inPorts []int

	outIP []net.IP

	outPorts [][]int

	inWorkers  Worker
	outWorkers []Worker
}

// Session is src or dst for workers
type Session struct {
	IP net.IP

	Ports []int
}

var (
	errSetSessError  = errors.New("Set session failed")
	errNodeBadInput  = errors.New("Bad input for node")
	errPipeExists    = errors.New("Pipe already exists")
	errPipeNotExists = errors.New("Pipe doesnot exists")
)

// Create a svr
func (sr *Node) Create() {
}

// AllocPull alloc one pull
func (sr *Node) AllocPull(id int, w Worker) error {
	var p *pipe

	if p = sr.all[id]; p == nil {
		p = &pipe{inPorts: []int{8000 + sr.Prefix + 2*id, 8000 + sr.Prefix + 2*id + 2}}
		sr.all[id] = p
	}

	if w == nil || !IsWorkerDec(w) {
		return errNodeBadInput
	}

	for _, exists := range p.outWorkers {
		if exists == w {
			return errPipeExists
		}
	}

	p.outWorkers = append(p.outWorkers, w)

	wid := GetWorkerWorkerID(w)

	ses := Session{Ports: []int{8000 + 2*wid, 8000 + 2*wid + 2}}
	if err := SetDecodeSes(w, &ses); err != nil {
		return err
	}

	// TODO: start here

	return nil
}

// FreePull free one pull
func (sr *Node) FreePull(id int, w Worker) error {
	var p *pipe
	if p = sr.all[id]; p == nil {
		return errPipeNotExists
	}

	if w == nil || !IsWorkerDec(w) {
		return errNodeBadInput
	}

	var exists Worker
	var k int
	for k, exists = range p.outWorkers {
		if exists == w {
			p.outWorkers[k] = nil
			break
		}
	}

	if exists != w {
		return errPipeNotExists
	}

	// ip := GetWorkerWorkerIP(w)
	// wid := GetWorkerWorkerID(w)

	// ports := []int{8000 + 2*wid, 8000 + 2*wid + 2}

	// TODO: free here

	return nil
}

// AllocPush return Pipe
func (sr *Node) AllocPush(id int, w Worker) error {
	var p *pipe

	if p = sr.all[id]; p == nil {
		p = &pipe{inPorts: []int{8000 + sr.Prefix + 4*id, 8000 + sr.Prefix + 4*id + 2}}
		sr.all[id] = p
	}

	if w == nil || !IsWorkerEnc(w) {
		return errNodeBadInput
	}

	if exists := p.inWorkers; exists != nil {
		// TODO re-do
	}

	ses := Session{IP: sr.IP,
		Ports: p.inPorts}

	if err := SetEncodeSes(w, &ses); err != nil {
		return errSetSessError
	}

	p.inWorkers = w

	return nil
}

// FreePush free one push
func (sr *Node) FreePush(id int, w Worker) error {
	var p *pipe
	if p = sr.all[id]; p == nil {
		return errPipeNotExists
	}

	if w == nil || !IsWorkerEnc(w) {
		return errNodeBadInput
	}

	return nil
}
