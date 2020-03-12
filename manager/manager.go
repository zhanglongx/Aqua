// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

// Package manager is the core module in Aqua, deals with
// config loading, sub-card's worker setting
package manager

import (
	"errors"
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
}

// Manager is main struct for mananger operation
type Manager struct {
	lock sync.RWMutex

	// DB store config settings
	DB DB

	// Workers store all sub-card's Workers
	Workers Workers
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

// Start does registing, and loads cfg from file
func (m *Manager) Start(DBFile string) error {

	m.Workers = Workers{}
	if err := m.Workers.register(); err != nil {
		return err
	}

	if err := m.DB.loadFromFile(DBFile); err != nil {
		return err
	}

	for path, params := range m.DB.Store {
		if err := m.Set(path, params); err != nil {
			comm.Error.Printf("appling saved params failed")

			// TODO: improve
			if err := m.DB.clearDB(); err != nil {
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

	w := m.Workers.findWorker(params.WorkerName)
	if w == nil {
		return errWorkerNotExists
	}

	if inUse := m.isWorkerAlloc(w); inUse != "" && inUse != path {
		return errWorkerInUse
	}

	if driver.IsWorkerDec(w) {
		ir, err := m.upstreamRes(params.UpStream)
		if err != nil {
			return err
		}

		// TODO: rtsp

		err = driver.SetDecodeRes(w, ir)
		if err != nil {
			return err
		}
	}

	if driver.IsWorkerEnc(w) {
		// TODO: rtsp
	}

	if err := driver.SetWorkerRunning(w, params.IsRunning); err != nil {
		return err
	}

	dupParams := *params
	if err := m.DB.set(path, &dupParams); err != nil {
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

	saved := m.DB.get(path)
	if saved == nil {
		// TODO: empty path?
		return Params{}, errPathNotExists
	}

	return *saved, nil
}

func (m *Manager) unAllocedWorkers() []string {

	var unUsed []string
	for _, w := range m.Workers {
		if m.isWorkerAlloc(w) == "" {
			unUsed = append(unUsed, driver.GetWorkerName(w))
		}
	}

	return unUsed
}

func (m *Manager) upstreamRes(up string) (driver.InnerRes, error) {

	if isPathValid(up) != nil {
		return driver.InnerRes{}, errPathNotExists
	}

	saved := m.DB.get(up)
	if saved == nil {
		return driver.InnerRes{}, errPathNotExists
	}

	upWorker := m.Workers.findWorker(saved.WorkerName)
	if upWorker == nil {
		return driver.InnerRes{}, errWorkerNotExists
	}

	return driver.GetEncodeRes(upWorker)
}

func (m *Manager) isWorkerAlloc(w driver.Worker) string {

	for path, params := range m.DB.Store {
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
