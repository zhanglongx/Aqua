// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

// Package manager is the core module in Aqua, deals with
// config loading, sub-card's worker setting and getting
package manager

import (
	"errors"
	"io"
	"regexp"
	"sort"
	"strconv"
	"sync"

	"github.com/xlab/treeprint"
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

	// workers store all workers can be assigned
	workers Workers
}

var (
	errBadParams       = errors.New("Params parse error")
	errPathNotExists   = errors.New("Path not exists")
	errWorkerNotExists = errors.New("Worker not exists")
	errWorkerInUse     = errors.New("Worker in Use")
)

// Create does registing, and loads cfg from file
func (ep *Path) Create(dir string, file string, need []string) error {

	ep.inUse = make(map[int]driver.Worker)

	ep.workers = Workers{}
	if err := ep.workers.register(need); err != nil {
		return err
	}

	ep.db.create()
	if err := ep.db.loadFromFile(dir, file); err != nil {
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
		if driver.IsWorkerDec(exists) {
			pipe := driver.Pipes[driver.PipeEncoder]
			if err := pipe.FreePull(ID, exists); err != nil {
				return err
			}
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

	if driver.IsWorkerDec(w) {
		pipe := driver.Pipes[driver.PipeEncoder]
		if err := pipe.AllocPull(ID, w); err != nil {
			return err
		}
	}

	if driver.IsWorkerEnc(w) {
		pipe := driver.Pipes[driver.PipeEncoder]
		if err := pipe.AllocPush(ID, w); err != nil {
			return err
		}
	}

	ep.inUse[ID] = w

	if card, ok := params["Card"].(map[string]interface{}); ok {
		if err := driver.SetWorkerSettings(w, card); err != nil {
			return err
		}
	} else {
		comm.Error.Printf("Param card format error")
	}

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

	sort.Strings(all)

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

// GetPipeInfo return a Pipesvr info
func GetPipeInfo(w io.Writer) {
	for k := range []int{driver.PipeRTSPIN, driver.PipeEncoder} {
		for _, p := range driver.Pipes[k].GetInfo() {

			tree := treeprint.New()
			var str string
			if p.InWorkers == nil {
				str = ""
			} else {
				str = driver.GetWorkerName(p.InWorkers)
			}
			node := tree.AddBranch(str)

			for _, o := range p.OutWorkers {
				if o != nil {
					node.AddNode(driver.GetWorkerName(o))
				}
			}

			w.Write([]byte(tree.String()))
		}
	}
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
