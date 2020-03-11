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

// Workers store all workers registered by cards
type Workers map[int][]driver.Worker

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

// R is the instance of reg
var R reg

// Init create M and R
func Init() {
	M = Manager{}

	R = reg{}
}

// Start does registing, and loads cfg from file
func (m *Manager) Start(DBFile string) error {

	var err error
	if m.Workers, err = R.Register(); err != nil {
		return err
	}

	if err := m.DB.loadFromFile(DBFile); err != nil {
		return err
	}

	for path, params := range m.DB.Store {
		if err := m.Set(path, params); err != nil {
			// TODO: improve

			comm.Error.Printf("appling saved params failed")
			m.DB.Store = make(map[string]*Params)
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

	if m.isWorkerAlloc(w) != path {
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
	m.DB.set(path, &dupParams)
	m.DB.saveToFile()

	return nil
}

// Get queries data
func (m *Manager) Get(path string) (Params, error) {

	m.lock.Lock()

	defer m.lock.Unlock()

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

func (w *Workers) findWorker(name string) driver.Worker {

	var s, i int
	for s = range *w {
		for i = range (*w)[s] {
			if name == driver.GetWorkerName((*w)[s][i]) {
				return (*w)[s][i]
			}
		}
	}

	return nil
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
