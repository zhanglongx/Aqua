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
	"regexp"
	"strconv"
	"sync"

	"github.com/zhanglongx/Aqua/driver"
)

// STR defines for data
const (
	STRPATH     = "PathName"
	STRWORKER   = "Worker"
	STRUPSTREAM = "UpStream"
	STRRUN      = "IsRunning"
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

	if driver.IsWorkerDec(worker) {
		ir, err := m.upstreamRes(data[STRUPSTREAM])
		if err != nil {
			return err
		}

		// TODO: rtsp

		err = driver.SetDecodeRes(worker, ir)
		if err != nil {
			return err
		}
	}

	if driver.IsWorkerEnc(worker) {
		// TODO: rtsp
	}

	var isRunning bool
	if isRunning {
		isRunning = true
	} else {
		isRunning = false
	}

	driver.SetWorkerRunning(worker, isRunning)

	// FIXME:
	re := regexp.MustCompile(`^_\d`)
	cardname := fmt.Sprintf("%q\n", re.Find([]byte(driver.GetWorkerName(worker))))

	ip := driver.GetWorkerIP(worker)

	rowDB := &pathRow{
		PathName:  data[STRPATH],
		Slot:      slot,
		WorkerID:  wid,
		CardName:  cardname,
		IP:        ip,
		IsRunning: isRunning,
		UpStream:  data[STRUPSTREAM],
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
	data[STRRUN] = getPathRunning(rowDB)

	// TODO: rtsp

	return data, nil
}

func (m *Manager) upstreamRes(up string) (driver.InnerRes, error) {

	if err := isPathValid(up); err != nil {
		return driver.InnerRes{}, err
	}

	rowDB := m.DB.get(up)
	if rowDB == nil {
		return driver.InnerRes{}, errPathNotExists
	}

	upWorker := m.Workers[rowDB.Slot][rowDB.WorkerID]
	if upWorker == nil {
		return driver.InnerRes{}, errPathNotExists
	}

	return driver.GetEncodeRes(upWorker)
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

func isPathValid(p string) error {
	if _, err := strconv.Atoi(p); err != nil {
		return err
	}

	return nil
}
