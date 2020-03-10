// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

// Package manager is the core module in Aqua, deals with
// config loading, sub-card's worker setting
package manager

import (
	"errors"
	"strconv"
	"strings"
	"sync"

	"github.com/zhanglongx/Aqua/driver"
)

// STR defines for data
const (
	STRPATH   = "PathName"
	STRWORKER = "Worker"
	STRRES    = "Res"
	STRRUN    = "IsRunning"
)

// InValidPathID is ID of Invalidate
const InValidPathID string = ""

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
	errPathNotExists   = errors.New("Path not exists")
	errWorkerNotExists = errors.New("Worker not exists")
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

// Start deals with initialization
func (m *Manager) Start(DBFile string) error {
	if err := m.DB.loadFromFile(DBFile); err != nil {
		return errors.New("load from file failed")
	}

	var err error
	if m.Workers, err = R.Register(); err != nil {
		return errors.New("register failed")
	}

	return nil
}

// Set processes data settings
func (m *Manager) Set(path string, data map[string]string) error {

	m.lock.Lock()

	defer m.lock.Unlock()

	if isPathValid(path) != nil {
		return errPathNotExists
	}

	worker, slot, wid := m.Workers.lookupWorker(data[STRWORKER])
	if worker == nil {
		return errWorkerNotExists
	}

	rowDB := &pathRow{
		Slot:     slot,
		WorkerID: wid,
	}

	m.DB.set(path, rowDB)

	return nil
}

// Get queries data
func (m *Manager) Get(path string) (map[string]string, error) {

	m.lock.Lock()

	defer m.lock.Unlock()

	if isPathValid(path) != nil {
		return nil, errPathNotExists
	}

	rowDB := m.DB.get(path)
	if rowDB == nil {
		// TODO: empty path?
		return nil, errPathNotExists
	}

	data := make(map[string]string)

	w := m.Workers[rowDB.Slot][rowDB.WorkerID]

	data[STRPATH] = getPathName(rowDB)
	data[STRWORKER] = driver.GetWorkerName(w)
	// tempz RES
	data[STRRUN] = getPathRunning(rowDB)

	return data, nil
}

func (w *Workers) lookupWorker(name string) (driver.Worker, int, int) {

	var s, i int
	for s = range *w {
		for i = range (*w)[s] {
			if name == driver.GetWorkerName((*w)[s][i]) {
				return (*w)[s][i], s, i
			}
		}
	}

	return nil, s, i
}

func (w *Workers) newRES(r string) driver.Resource {
	if strings.HasPrefix(strings.ToUpper(r), "rtsp://") {
		return driver.OutterRes{Rtsp: r}
	}

	if _, ok := strconv.Atoi(r); ok != nil {
		return nil
	}

	return nil
}

func isPathValid(p string) error {
	if _, err := strconv.Atoi(p); err != nil {
		return err
	}

	return nil
}
