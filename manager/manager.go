// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

// Package manager is the core module in Aqua, deals with
// config loading, sub-card's worker setting and getting
package manager

import (
	"errors"
	"regexp"
	"strconv"
	"sync"

	"github.com/zhanglongx/Aqua/comm"
	"github.com/zhanglongx/Aqua/driver"
)

// Params is the main data struct used to set and
// get path setttings
type Params map[string]interface{}

// Path is the main struct for control sub-cards
type Path struct {
	lock sync.RWMutex

	// db store settings
	db DB

	// inUse holds all in use driver.Worker
	inUse map[int]driver.Worker

	// workers store all sub-card's workers
	workers Workers
}

var (
	errBadParams       = errors.New("Params parse error")
	errPathNotExists   = errors.New("Path not exists")
	errWorkerNotExists = errors.New("Worker not exists")
	errWorkerInUse     = errors.New("Worker in Use")
)

// EPath is the instance of EncoderPath
var EPath Path = Path{}

// Create does registing, and loads cfg from file
func (ep *Path) Create(DBFile string) error {

	ep.inUse = make(map[int]driver.Worker)

	// tempz: receive from Parameters
	ep.workers = Workers{}
	if err := ep.workers.register(); err != nil {
		return err
	}

	ep.db.create()
	if err := ep.db.loadFromFile(DBFile); err != nil {
		return err
	}

	for IDStr, params := range ep.db.Params {
		id, _ := strconv.Atoi(IDStr)

		if err := ep.Set(id, params); err != nil {
			comm.Error.Printf("Appling saved params in path %d failed", id)

			// Just clear the path?
			if err := ep.db.set(id, nil); err != nil {
				return err
			}
		}
	}

	return nil
}

// Set processes data settings
func (ep *Path) Set(ID int, params Params) error {

	ep.lock.Lock()

	defer ep.lock.Unlock()

	if !isPathValid(ID) {
		return errPathNotExists
	}

	if err := checkParams(params); err != nil {
		return err
	}

	w := ep.workers.findWorker(params["WorkerName"].(string))
	if w == nil {
		return errWorkerNotExists
	}

	if k := ep.isWorkerAlloc(w); k != -1 && k != ID {
		return errWorkerInUse
	}

	if exists := ep.inUse[ID]; exists != nil {
		// un-do
		// FIXME: only true in EncodePath
		if driver.IsWorkerDec(exists) {
			pipe := driver.Pipes[driver.PipeRTSPIN]
			if err := pipe.FreePush(ID); err != nil {
				return err
			}

			if err := pipe.FreePull(ID, ep.inUse[ID]); err != nil {
				return err
			}

			// hack: double free to be sure pipe.Pull is freed anyway
			pipe = driver.Pipes[driver.PipeEncoder]
			if err := pipe.FreePull(ID, ep.inUse[ID]); err != nil {
				return err
			}

			// FIXME: un-do RTSPIn ?
		}

		if driver.IsWorkerEnc(exists) {
			pipe := driver.Pipes[driver.PipeEncoder]
			if err := pipe.FreePush(ID); err != nil {
				return err
			}
		}

		// TODO: maybe more?

		ep.inUse[ID] = nil
	}

	// RTSP
	if params["RtspIn"].(string) != "" {
		// hack: if it's a rtsp worker
		if driver.IsWorkerDec(w) {
			rtsp := ep.workers.findWorker("rtsp_254_0")
			if rtsp == nil {
				return errWorkerNotExists
			}

			// TODO: control rtsp in

			pipe := driver.Pipes[driver.PipeRTSPIN]
			if err := pipe.AllocPush(ID, rtsp); err != nil {
				return err
			}

			if err := pipe.AllocPull(ID, w); err != nil {
				return err
			}
		} else {
			// TODO: rtsp direct in
		}
	} else {
		// Inner
		if driver.IsWorkerDec(w) {
			pipe := driver.Pipes[driver.PipeEncoder]
			if err := pipe.AllocPull(ID, w); err != nil {
				return err
			}
		}
	}

	if driver.IsWorkerEnc(w) {
		pipe := driver.Pipes[driver.PipeEncoder]
		if err := pipe.AllocPush(ID, w); err != nil {
			return err
		}

		// TODO: rtspOut
		// rtsp := ep.workers.findWorker("rtsp_255_0")
		// if rtsp == nil {
		// 	return errWorkerNotExists
		// }

		// if err := pipe.AllocPull(ID, rtsp); err != nil {
		// 	return err
		// }
	}

	ep.inUse[ID] = w

	// TODO: apply params to Workers

	isRunning := params["IsRunning"].(bool)
	if err := driver.SetWorkerRunning(w, isRunning); err != nil {
		return err
	}

	if err := ep.db.set(ID, params); err != nil {
		return err
	}

	return nil
}

// Get queries data
func (ep *Path) Get(ID int) (Params, error) {

	ep.lock.RLock()

	defer ep.lock.RUnlock()

	if !isPathValid(ID) {
		return nil, errPathNotExists
	}

	saved := ep.db.get(ID)
	if saved == nil {
		// TODO: empty path?
		return nil, errPathNotExists
	}

	return saved, nil
}

// GetWorkers gets all workers registered under a path
func (ep *Path) GetWorkers() []string {

	ep.lock.RLock()

	defer ep.lock.RUnlock()

	var all []string
	for _, w := range ep.workers {
		all = append(all, driver.GetWorkerName(w))
	}

	return all
}

// isWorkerAlloc find if a worker is alloc
func (ep *Path) isWorkerAlloc(w driver.Worker) int {
	for k, exist := range ep.inUse {
		if exist == w {
			return k
		}
	}

	return -1
}

func isPathValid(ID int) bool {
	if ID < 0 {
		return false
	}

	return true
}

// checkParams only do basic literal check, and leaves legal
// checking alone
func checkParams(params Params) error {

	if params == nil {
		// TODO: un-do a path?
		return errBadParams
	}

	// TODO: unicode
	wn := params["WorkerName"].(string)
	matched, err := regexp.Match(`\S+_\d+_\d+`, []byte(wn))
	if !matched || err != nil {
		return errBadParams
	}

	return nil
}
