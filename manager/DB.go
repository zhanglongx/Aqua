// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

// Package manager deals with
package manager

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
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

	// sub-card's worker ID
	WorkerID driver.WorkerID

	// sub-card name
	Name driver.NameID

	// sub-card IP
	IP driver.IP

	// path's status
	IsRunning driver.IsRunning

	// input resource
	InRes []driver.Resource
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
		d.Config = make(map[pathID]*pathRow, 0)
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

// query queries pathID on pathRow, base on Slot, WorkerID
// IP and return pathID
func (d *DB) query(p *pathRow) pathID {
	if p == nil {
		return InValidPathID
	}

	d.lock.RLock()

	defer d.lock.RUnlock()

	for k, c := range d.Config {
		if c.Slot == p.Slot &&
			c.WorkerID == p.WorkerID &&
			c.Name == p.Name &&
			net.IP(c.IP).Equal(net.IP(p.IP)) {

			// same as previous
			return k
		}
	}

	return InValidPathID
}

// set set a new pathRow in DB, DON'T reuse
// *pathRow return by get
func (d *DB) set(ID pathID, p *pathRow) error {
	d.lock.Lock()

	defer d.lock.Unlock()

	d.Config[ID] = p

	return nil
}

// get return a pathRow in DB, DON'T reuse
// *pathRow return by set
func (d *DB) get(ID pathID) *pathRow {
	d.lock.RLock()

	defer d.lock.RUnlock()

	if d.Config[ID] == nil {
		return nil
	}

	return d.Config[ID]
}
