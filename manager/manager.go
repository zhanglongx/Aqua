// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

// Package manager is the core module in Aqua, deals with
// config loading, sub-card's worker setting
package manager

import (
	"errors"
	"fmt"
	"sync"

	"github.com/zhanglongx/Aqua/driver"
)

// STRWORKER for quering
const STRWORKER = "Worker"

// STRPATH for quering
const STRPATH = "PathName"

// STRRES for quering
const STRRES = "Res"

// STRRUN for quering
const STRRUN = "IsRunning"

// pathID is the path's ID
type pathID string

// InValidPathID is ID of Invalidate
const InValidPathID pathID = ""

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
func (m *Manager) Set(path pathID, data map[string]string) error {

	m.lock.Lock()

	defer m.lock.Unlock()

	return nil
}

// Get queries data
func (m *Manager) Get(path pathID) (map[string]string, error) {

	m.lock.Lock()

	defer m.lock.Unlock()

	rowDB := m.DB.get(path)
	if rowDB == nil {
		return make(map[string]string), nil
	}

	data := make(map[string]string)

	// XXX: make sure rowDB copied before return
	m.getWorker(data, rowDB.Slot, rowDB.WorkerID)

	data[STRPATH] = rowDB.PathName
	data[STRRES] = fmt.Sprintf("%v", rowDB.InRes[0]) // tempz
	data[STRRUN] = fmt.Sprintf("%v", rowDB.IsRunning)

	return data, nil
}

// getWorker return generic worker's info
func (m *Manager) getWorker(data map[string]string, s int, w int) {

	worker := m.Workers[s][w]

	data[STRWORKER] = fmt.Sprintf("%v", worker)
}
