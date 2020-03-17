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
	errPipeNotExists = errors.New("Pipe does not exists")
)

// Create a svr
func (n *Node) Create() {
	n.all = make(map[int]*pipe)
}

// AllocPull alloc one pull
func (n *Node) AllocPull(id int, w Worker) error {
	var p *pipe

	if p = n.all[id]; p == nil {
		p = &pipe{inPorts: []int{8000 + n.Prefix + 2*id, 8000 + n.Prefix + 2*id + 2}}
		n.all[id] = p
	}

	if w == nil || !IsWorkerDec(w) {
		return errNodeBadInput
	}

	for _, exists := range p.outWorkers {
		if exists == w {
			return nil
		}
	}

	wid := GetWorkerWorkerID(w)
	// IP := GetWorkerWorkerIP(w)

	ses := Session{Ports: []int{8000 + 2*wid, 8000 + 2*wid + 2}}
	if err := SetDecodeSes(w, &ses); err != nil {
		return err
	}

	// TODO: start here

	p.outWorkers = append(p.outWorkers, w)

	return nil
}

// FreePull free one pull
func (n *Node) FreePull(id int, w Worker) error {
	var p *pipe
	if p = n.all[id]; p == nil {
		return errPipeNotExists
	}

	if w == nil || !IsWorkerDec(w) {
		return errNodeBadInput
	}

	var exists Worker
	var k int
	for k, exists = range p.outWorkers {
		if exists == w {
			break
		}
	}

	if exists != w {
		return errPipeNotExists
	}

	// TODO: free here

	p.outWorkers[k] = nil

	return nil
}

// AllocPush alloc one push
func (n *Node) AllocPush(id int, w Worker) error {
	var p *pipe

	if p = n.all[id]; p == nil {
		p = &pipe{inPorts: []int{8000 + n.Prefix + 4*id, 8000 + n.Prefix + 4*id + 2}}
		n.all[id] = p
	}

	if w == nil || !IsWorkerEnc(w) {
		return errNodeBadInput
	}

	if exists := p.inWorkers; exists != nil {
		if exists == w {
			return nil
		}
		// TODO re-do
	}

	ses := Session{IP: n.IP,
		Ports: p.inPorts}

	if err := SetEncodeSes(w, &ses); err != nil {
		return errSetSessError
	}

	p.inWorkers = w

	return nil
}

// FreePush free one push
func (n *Node) FreePush(id int, w Worker) error {
	var p *pipe
	if p = n.all[id]; p == nil {
		return errPipeNotExists
	}

	if w == nil || !IsWorkerEnc(w) {
		return errNodeBadInput
	}

	// TODO: free here

	p.inWorkers = nil

	return nil
}
