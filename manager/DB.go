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
	Store map[string]*Params
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
	d.Store = make(map[string]*Params, 0)
	return d.saveToFile()
}

// set set a new pathRow in DB, DB store *ONLY* the pointer
// passed in, so make sure passing a whole new pathRow{}
// everytime
func (d *DB) set(ID string, p *Params) error {

	d.Store[ID] = p

	return d.saveToFile()
}

// get return a *pathRow in DB. Because set() and get() are
// not thread-safe, you should get data in *pathRow copied,
// before using any unlock method
func (d *DB) get(ID string) *Params {

	if d.Store[ID] == nil {
		return nil
	}

	return d.Store[ID]
}
