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
type Params map[string]interface{}

// EncodePath is the main struct for Encoder's Path
type EncodePath struct {
	lock sync.RWMutex

	// db store settings
	db DB

	// encoders holds all driver.Encoder
	encoders map[int]driver.Worker

	// workers store all sub-card's workers
	workers Workers
}

var (
	errBadParams       = errors.New("Params parse error")
	errPathNotExists   = errors.New("Path not exists")
	errWorkerNotExists = errors.New("Worker not exists")
	errWorkerInUse     = errors.New("Worker In Use")
)

// EPath is the instance of EncoderPath
var EPath EncodePath = EncodePath{}

// Create does registing, and loads cfg from file
func (ep *EncodePath) Create(DBFile string) error {

	ep.encoders = make(map[int]driver.Worker)

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
func (ep *EncodePath) Set(ID int, params Params) error {

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

	if k := ep.workers.isWorkerAlloc(w); k != -1 && k != ID {
		return errWorkerInUse
	}

	if ep.encoders[ID] != nil {
		// un-do
		pipe := driver.Pipes[driver.PipeRTSPIN]
		if err := pipe.FreePush(ID); err != nil {
			return err
		}

		if err := pipe.FreePull(ID, ep.encoders[ID]); err != nil {
			return err
		}

		pipe = driver.Pipes[driver.PipeEncoder]
		if err := pipe.FreePush(ID); err != nil {
			return err
		}

		ep.encoders[ID] = nil
	}

	// RTSP
	if driver.IsWorkerDec(w) {
		if params["RtspIn"].(string) != "" {
			rtsp := ep.workers.findWorker("rtsp_254_0")
			if rtsp == nil {
				return errWorkerNotExists
			}

			pipe := driver.Pipes[driver.PipeRTSPIN]
			if err := pipe.AllocPush(ID, rtsp); err != nil {
				return err
			}

			if err := pipe.AllocPull(ID, w); err != nil {
				return err
			}
		}
	}

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

	ep.encoders[ID] = w

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
func (ep *EncodePath) Get(ID int) (Params, error) {

	ep.lock.RLock()

	defer ep.lock.RUnlock()

	if !isPathValid(ID) {
		return Params{}, errPathNotExists
	}

	saved := ep.db.get(ID)
	if saved == nil {
		// TODO: empty path?
		return Params{}, errPathNotExists
	}

	return saved, nil
}

func unUsedWorkers(inUse []driver.Worker, ws Workers) []string {

	var unUsed []string
	for _, w := range inUse {
		if w == nil {
			continue
		}

		if ws.isWorkerAlloc(w) == -1 {
			unUsed = append(unUsed, driver.GetWorkerName(w))
		}
	}

	return unUsed
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
