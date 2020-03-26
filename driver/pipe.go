// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

package driver

import (
	"errors"
	"net"
	"sync"
)

const (
	inBasePort  = 6000
	outBasePort = 6000
)

// PipeSvr alloc Pipe
type PipeSvr struct {
	lock sync.Mutex

	// IP is the Svr IP
	IP net.IP

	// Prefix to identity services
	Prefix int

	all map[int]*Pipe
}

// Pipe contains pipeline info used by PipeSvr
type Pipe struct {
	inPorts []int

	outIP []net.IP

	outPorts [][]int

	InWorkers  Worker
	OutWorkers []Worker
}

// Session is src or dst for workers
type Session struct {
	IP net.IP

	Ports []int
}

var (
	errNodeBadInput = errors.New("Bad input for node")
)

func helperPort(base int, prefix int, id int) []int {
	return []int{base + prefix + 4*id, base + prefix + 4*id + 2}
}

// Create a svr
func (sr *PipeSvr) Create() {
	sr.all = make(map[int]*Pipe)
}

// AllocPull alloc one pull
func (sr *PipeSvr) AllocPull(id int, w Worker) error {
	var p *Pipe

	sr.lock.Lock()

	defer sr.lock.Unlock()

	if p = sr.all[id]; p == nil {
		p = &Pipe{inPorts: helperPort(inBasePort, sr.Prefix, id)}
		sr.all[id] = p
	}

	if w == nil || !IsWorkerDec(w) {
		return errNodeBadInput
	}

	for _, exists := range p.OutWorkers {
		if exists == w {
			return nil
		}
	}

	wid := GetWorkerWorkerID(w)
	// IP := GetWorkerWorkerIP(w)

	ses := Session{Ports: helperPort(outBasePort, sr.Prefix, wid)}
	if err := SetDecodeSes(w, &ses); err != nil {
		return err
	}

	// TODO: start here

	p.OutWorkers = append(p.OutWorkers, w)

	return nil
}

// FreePull free one pull
func (sr *PipeSvr) FreePull(id int, w Worker) error {
	var p *Pipe

	sr.lock.Lock()

	defer sr.lock.Unlock()

	if p = sr.all[id]; p == nil {
		return nil
	}

	if w == nil || !IsWorkerDec(w) {
		return errNodeBadInput
	}

	var k int
	var exists Worker
	for k, exists = range p.OutWorkers {
		if exists == w {
			break
		}
	}

	if exists != w {
		return nil
	}

	// TODO: free here

	p.OutWorkers = remove(p.OutWorkers, k)

	return nil
}

// AllocPush alloc one push
func (sr *PipeSvr) AllocPush(id int, w Worker) error {

	sr.lock.Lock()

	defer sr.lock.Unlock()

	var p *Pipe

	if p = sr.all[id]; p == nil {
		p = &Pipe{inPorts: helperPort(inBasePort, sr.Prefix, id)}
		sr.all[id] = p
	}

	if w == nil || !IsWorkerEnc(w) {
		return errNodeBadInput
	}

	if exists := p.InWorkers; exists != nil {
		if exists == w {
			return nil
		}
		// TODO: un-do
	}

	ses := Session{IP: sr.IP, Ports: p.inPorts}

	if err := SetEncodeSes(w, &ses); err != nil {
		return err
	}

	// TODO: push here

	p.InWorkers = w

	return nil
}

// FreePush free one push
func (sr *PipeSvr) FreePush(id int) error {

	sr.lock.Lock()

	defer sr.lock.Unlock()

	var p *Pipe

	if p = sr.all[id]; p == nil {
		return nil
	}

	// TODO: free here

	p.InWorkers = nil

	return nil
}

// GetInfo print tree-like string
func (sr *PipeSvr) GetInfo() []Pipe {

	sr.lock.Lock()

	defer sr.lock.Unlock()

	var out []Pipe

	for _, p := range sr.all {
		if p.InWorkers == nil && len(p.OutWorkers) == 0 {
			continue
		}

		out = append(out, *p)
	}

	return out
}

// https://yourbasic.org/golang/delete-element-slice/
func remove(ws []Worker, i int) []Worker {
	// Remove the element at index i from a.
	ws[i] = ws[len(ws)-1] // Copy last element to index i.
	ws[len(ws)-1] = nil   // Erase last element (write zero value).
	ws = ws[:len(ws)-1]   // Truncate slice.

	return ws
}
