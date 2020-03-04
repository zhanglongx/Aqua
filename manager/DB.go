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
// set() and get() are not thread-safe, it's caller's
// responsibility to ensure that. To ensure data returned by
// get() will not get rewritten, data from pathRow should be
// copied before any unlock method.
type DB struct {
	// Version should be used to check DB's compatibility
	Version string

	// Config stores all the configurations
	Config map[pathID]*pathRow
}

var errJSONFILE = errors.New("DB: JSON File error")
var errPathExists = errors.New("DB: path already exists")

// loadFromFile load JSON file to Cfg
func (d *DB) loadFromFile(JFile string) error {

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
		comm.Error.Printf("Discarding old file: %s\n", JFile)
		d.Config = make(map[pathID]*pathRow, 0)
		return nil
	}

	// all pathDB.validate will be set to false
	return nil
}

// saveToFile save JSON file to Cfg
func (d *DB) saveToFile(JFile string) error {

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

// query queries pathID on pathRow, if Slot, WorkerID, Name and IP
// are identity, return pathID
func (d *DB) query(p *pathRow) pathID {

	if p == nil {
		return InValidPathID
	}

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

// set set a new pathRow in DB, DB store *ONLY* the pointer
// passed in, so make sure passing a whole new pathRow{}
// everytime
func (d *DB) set(ID pathID, p *pathRow) error {

	d.Config[ID] = p

	return nil
}

// get return a *pathRow in DB. Because set() and get() are
// not thread-safe, you should get data in *pathRow copied,
// before using any unlock method
func (d *DB) get(ID pathID) *pathRow {

	if d.Config[ID] == nil {
		return nil
	}

	return d.Config[ID]
}
