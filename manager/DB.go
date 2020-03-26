// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

package manager

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/zhanglongx/Aqua/comm"
)

// DBVER is DB File Version
const DBVER string = "1.0.0"

// DB contains all path' config. It's degsinged to be easily
// exported to file (like JSON).
// set() and get() are not thread-safe, it's caller's
// responsibility to ensure that.
// It's designed to copy params in set() and get()
type DB struct {
	fullPathFile string

	// Version should be used to check DB's compatibility
	Version string

	// Store contains all path params
	Params map[string]Params
}

// create initialize Params
func (d *DB) create() {
	d.Params = make(map[string]Params)
}

// loadFromFile load JSON file to Cfg
func (d *DB) loadFromFile(dir string, file string) error {

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		comm.Error.Printf("Dir %s not exists, create", dir)
		if err := os.Mkdir(dir, 0644); err != nil {
			return err
		}
	}

	fullPathFile := path.Join(dir, file)
	if fullPathFile == "" {
		comm.Error.Fatalf("Join %s %s failed", dir, file)
		return nil
	}

	if _, err := os.Stat(fullPathFile); err != nil {
		if os.IsNotExist(err) {

			comm.Info.Printf("DB %s not exists, create", fullPathFile)

			d.Version = DBVER
			d.fullPathFile = fullPathFile
			return nil
		}

		return err
	}

	buf, err := ioutil.ReadFile(fullPathFile)
	if err != nil {
		comm.Error.Printf("Read DB file %s failed", fullPathFile)
		return err
	}

	err = json.Unmarshal(buf, d)
	if err != nil {
		comm.Error.Printf("Decode DB file %s failed", fullPathFile)
		return err
	}

	// FIXME: more compatible
	if d.Version != DBVER {
		comm.Error.Printf("DB file ver error: %s", d.Version)
		comm.Error.Printf("Discarding old file: %s", fullPathFile)
		if err := d.clearDB(); err != nil {
			return err
		}

	}

	d.fullPathFile = fullPathFile

	return nil
}

// saveToFile save JSON file to Cfg
func (d *DB) saveToFile() error {

	buf, err := json.MarshalIndent(d, "", "    ")
	if err != nil {
		comm.Error.Printf("Encode DB %s failed", d.fullPathFile)
		return err
	}

	err = ioutil.WriteFile(d.fullPathFile, buf, 0644)
	if err != nil {
		comm.Error.Printf("Write DB file %s failed", d.fullPathFile)
		return err
	}

	return nil
}

// clearDB clear DB
func (d *DB) clearDB() error {
	comm.Info.Printf("Clearing DB: %s", d.fullPathFile)

	d.Params = make(map[string]Params)
	return d.saveToFile()
}

// set set a new Param in DB with informed ID
func (d *DB) set(ID int, p Params) error {

	if ID < 0 {
		return nil
	}

	id := fmt.Sprintf("%d", ID)

	if p == nil {
		delete(d.Params, id)
		return d.saveToFile()
	}

	d.Params[id] = p
	return d.saveToFile()
}

// get get a exist Param in DB with informed ID
func (d *DB) get(ID int) Params {

	if ID < 0 {
		return nil
	}

	id := fmt.Sprintf("%d", ID)

	return d.Params[id]
}
