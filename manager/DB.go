// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

// Package manager deals with
package manager

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"sync"

	"github.com/zhanglongx/Aqua/comm"
	"github.com/zhanglongx/Aqua/driver"
)

// DBVER is DB File Version
const DBVER string = "1.0.0"

// pathRow is the row-query struct.
type pathRow struct {
	// sub-card slot number
	Slot driver.SlotID

	// sub-card name
	Name driver.NameID

	// sub-card IP
	IP driver.IP

	// sub-card's worker ID
	WorkerID driver.WorkerID

	// path's status
	IsRunning driver.IsRunning

	// input resource
	InRes []driver.Resource

	// output resource
	OutRes driver.Resource
}

// DB contains all path' config. It's degsinged to be easily
// exported to file (like JSON).
type DB struct {
	lock sync.RWMutex

	Version string

	Config map[pathID]*pathRow
}

var errJSONFILE = errors.New("DB: JSON File error")
var errPathExists = errors.New("DB: path already exists")

// loadFromFile load JSON file to Cfg
func (d *DB) loadFromFile(JFile string) error {
	d.lock.Lock()

	defer d.lock.Unlock()

	buf, err := ioutil.ReadFile(JFile)
	if err != nil {
		comm.Error.Printf("Read DB file %s failed\n", JFile)
		return err
	}

	err = json.Unmarshal(buf, d)
	if err != nil {
		comm.Error.Printf("Decode DB file %s failed\n", JFile)
		return err
	}

	// FIXME: more compatible
	if d.Version != DBVER {
		comm.Error.Printf("DB file ver error: %s\n", d.Version)
		return errJSONFILE
	}

	// all pathDB.validate will be set to false
	return nil
}

// saveToFile save JSON file to Cfg
func (d *DB) saveToFile(JFile string) error {
	d.lock.RLock()

	defer d.lock.RUnlock()

	buf, err := json.Marshal(d)
	if err != nil {
		comm.Error.Printf("Encode DB %s failed\n", JFile)
		return err
	}

	err = ioutil.WriteFile(JFile, buf, 0644)
	if err != nil {
		comm.Error.Printf("Write DB file %s failed\n", JFile)
		return err
	}

	return nil
}

// query queries pathID on pathRow, and return path's IsRunning
// to caller
func (d *DB) query(p *pathRow) pathID {
	if p == nil {
		return InValidPathID
	}

	d.lock.RLock()

	defer d.lock.RUnlock()

	for k, c := range d.Config {
		if c.Slot == p.Slot &&
			c.Name == p.Name &&
			bytes.Compare(c.IP, p.IP) == 0 &&
			c.WorkerID == p.WorkerID {

			p.IsRunning = c.IsRunning
			p.InRes = c.InRes
			p.OutRes = c.OutRes
			return k
		}
	}

	return InValidPathID
}

// add add a new pathRow to DB
func (d *DB) add(ID pathID, p *pathRow) error {
	d.lock.Lock()

	defer d.lock.Unlock()

	if d.Config[ID] != nil {
		return errPathExists
	}

	d.Config[ID] = new(pathRow)
	*d.Config[ID] = *p

	return nil
}
