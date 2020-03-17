// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

package manager

import (
	"encoding/json"
	"io/ioutil"

	"github.com/zhanglongx/Aqua/comm"
)

// DBVER is DB File Version
const DBVER string = "1.0.0"

// Params is the main struct used to set and
// get path setttings
type Params map[string]interface{}

// DB contains all path' config. It's degsinged to be easily
// exported to file (like JSON).
// set() and get() are not thread-safe, it's caller's
// responsibility to ensure that. To ensure data returned by
// get() will not get rewritten, data from pathRow should be
// copied before any unlock method.
type DB struct {
	jFile string

	// Version should be used to check DB's compatibility
	Version string

	// Store contains all path params
	Params []Params
}

// create initialize Params
func (d *DB) create() {
	d.Params = make([]Params, 32)
}

// loadFromFile load JSON file to Cfg
func (d *DB) loadFromFile(JFile string) error {

	buf, err := ioutil.ReadFile(JFile)
	if err != nil {
		comm.Error.Printf("Read DB file %s failed", JFile)
		return err
	}

	err = json.Unmarshal(buf, d)
	if err != nil {
		comm.Error.Printf("Decode DB file %s failed", JFile)
		return err
	}

	// FIXME: more compatible
	if d.Version != DBVER {
		comm.Error.Printf("DB file ver error: %s", d.Version)
		comm.Error.Printf("Discarding old file: %s", JFile)
		if err := d.clearDB(); err != nil {
			return err
		}
		return nil
	}

	d.jFile = JFile

	return nil
}

// saveToFile save JSON file to Cfg
func (d *DB) saveToFile() error {

	buf, err := json.MarshalIndent(d, "", "    ")
	if err != nil {
		comm.Error.Printf("Encode DB %s failed", d.jFile)
		return err
	}

	err = ioutil.WriteFile(d.jFile, buf, 0644)
	if err != nil {
		comm.Error.Printf("Write DB file %s failed", d.jFile)
		return err
	}

	return nil
}

// clearDB clear DB
func (d *DB) clearDB() error {

	d.Params = make([]Params, 32)
	return d.saveToFile()
}

// set set a new Param in DB with informed ID
func (d *DB) set(ID int, p Params) error {

	if ID < 0 {
		return nil
	}

	if ID >= len(d.Params) {
		t := make([]Params, len(d.Params), (cap(d.Params)+1)*2)
		copy(t, d.Params)

		d.Params = t
	}

	d.Params[ID] = p
	return d.saveToFile()
}

// get get a exist Param in DB with informed ID
func (d *DB) get(ID int) Params {

	if ID < 0 {
		return nil
	}

	if ID >= len(d.Params) {
		return nil
	}

	return d.Params[ID]
}
