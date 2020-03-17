// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

// Package manager is the core module in Aqua, deals with
// config loading, sub-card's worker setting
package manager

import (
	"errors"
	"net"
	"regexp"
	"strconv"
	"sync"

	"github.com/zhanglongx/Aqua/comm"
	"github.com/zhanglongx/Aqua/driver"
)

// Params is the main struct used to set and
// get path setttings
type Params struct {
	// Path Name
	PathName string

	// Worker name
	WorkerName string

	// path's status
	IsRunning bool

	// UpStream pathID
	UpStream string

	// Rtsp in
	RtspIn string

	// Rtsp out
	RtspOut string
}

// Nodes ID
const (
	LeftNode = iota
	RightNode
	MaxNodes
)

// Manager is main struct for mananger operation
type Manager struct {
	lock sync.RWMutex

	// db store settings
	db DB

	// PipeSrv
	nodes [MaxNodes]driver.Node

	// workers store all sub-card's workers
	workers Workers
}

var (
	errBadParams       = errors.New("Params parse error")
	errPathNotExists   = errors.New("Path not exists")
	errWorkerNotExists = errors.New("Worker not exists")
	errWorkerInUse     = errors.New("Worker In Use")
)

// M is the instance of Manager
var M Manager

// Init create M and R
func Init() {
	M = Manager{}
}

// Init does registing, and loads cfg from file
func (m *Manager) Init(DBFile string) error {

	m.workers = Workers{}
	if err := m.workers.register(); err != nil {
		return err
	}

	if err := m.db.loadFromFile(DBFile); err != nil {
		return err
	}

	// tempz
	m.nodes[LeftNode] = driver.Node{IP: net.IPv4(192, 165, 53, 35),
		Prefix: 1000}
	m.nodes[RightNode] = driver.Node{IP: net.IPv4(192, 165, 53, 35),
		Prefix: 0}

	m.nodes[LeftNode].Create()
	m.nodes[RightNode].Create()

	for path, params := range m.db.Store {
		if err := m.Set(path, params); err != nil {
			comm.Error.Printf("Appling saved params in path %s failed",
				path)

			// TODO: improve
			if err := m.db.clearDB(); err != nil {
				return err
			}
			break
		}
	}

	return nil
}

// Set processes data settings
func (m *Manager) Set(path string, params *Params) error {

	m.lock.Lock()

	defer m.lock.Unlock()

	if isPathValid(path) != nil {
		return errPathNotExists
	}

	if err := checkParams(params); err != nil {
		return err
	}

	w := m.workers.findWorker(params.WorkerName)
	if w == nil {
		return errWorkerNotExists
	}

	if inUse := m.isWorkerAlloc(w); inUse != "" && inUse != path {
		return errWorkerInUse
	}

	if driver.IsWorkerDec(w) {
		id, _ := strconv.Atoi(path)

		if params.RtspIn != "" {
			// tempz
			rtsp := m.workers.findWorker("rtsp_254_0")
			if rtsp == nil {
				return errWorkerNotExists
			}

			if err := m.nodes[LeftNode].AllocPush(id, rtsp); err != nil {
				return err
			}

			if err := m.nodes[LeftNode].AllocPull(id, w); err != nil {
				return err
			}
		} else {
			if _, err := m.findUP(params.UpStream); err != nil {
				return err
			}

			id, _ = strconv.Atoi(params.UpStream)

			if err := m.nodes[RightNode].AllocPull(id, w); err != nil {
				return err
			}
		}
	}

	if driver.IsWorkerEnc(w) {
		id, _ := strconv.Atoi(path)

		if params.RtspOut != "" {
			// tempz
			rtsp := m.workers.findWorker("rtsp_255_0")
			if rtsp == nil {
				return errWorkerNotExists
			}

			if err := m.nodes[RightNode].AllocPull(id, rtsp); err != nil {
				return err
			}
		}

		if err := m.nodes[RightNode].AllocPush(id, w); err != nil {
			return err
		}
	}

	if err := driver.SetWorkerRunning(w, params.IsRunning); err != nil {
		return err
	}

	dupParams := *params
	if err := m.db.set(path, &dupParams); err != nil {
		return err
	}

	return nil
}

// Get queries data
func (m *Manager) Get(path string) (Params, error) {

	m.lock.RLock()

	defer m.lock.RUnlock()

	if isPathValid(path) != nil {
		return Params{}, errPathNotExists
	}

	saved := m.db.get(path)
	if saved == nil {
		// TODO: empty path?
		return Params{}, errPathNotExists
	}

	return *saved, nil
}

func (m *Manager) unAllocedWorkers() []string {

	var unUsed []string
	for _, w := range m.workers {
		if m.isWorkerAlloc(w) == "" {
			unUsed = append(unUsed, driver.GetWorkerName(w))
		}
	}

	return unUsed
}

func (m *Manager) findUP(up string) (driver.Worker, error) {

	if isPathValid(up) != nil {
		return nil, errPathNotExists
	}

	saved := m.db.get(up)
	if saved == nil {
		return nil, errPathNotExists
	}

	upWorker := m.workers.findWorker(saved.WorkerName)
	if upWorker == nil {
		return nil, errWorkerNotExists
	}

	return upWorker, nil
}

func (m *Manager) isWorkerAlloc(w driver.Worker) string {

	for path, params := range m.db.Store {
		if params.WorkerName == driver.GetWorkerName(w) {
			return path
		}
	}

	return ""
}

func isPathValid(p string) error {
	if _, err := strconv.Atoi(p); err != nil {
		return err
	}

	return nil
}

// checkParams only do basic literal check, and leaves legal
// checking alone
func checkParams(p *Params) error {

	// TODO: unicode
	matched, err := regexp.Match(`\S+_\d+_\d+`, []byte(p.WorkerName))
	if !matched || err != nil {
		return errBadParams
	}

	return nil
}
