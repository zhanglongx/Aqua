// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

// Package manager deals with
package manager

import (
	"errors"
	"sync"

	"github.com/zhanglongx/Aqua/driver"
)

// pathDB is the row-query struct.
type pathDB struct {
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

	// pathDB can be loaded, but not validate this time
	validate bool
}

// DB contains all path' config. It's degsinged to be easily
// exported to file (like JSON).
type DB struct {
	lock sync.RWMutex

	Version string

	Config map[pathID]*pathDB
}

var errJSONFILE = errors.New("cfg: JSON File Format error")

// LoadFromFile load JSON file to Cfg
func (c *DB) LoadFromFile(JFile string) error {
	c.lock.Lock()

	defer c.lock.Unlock()

	// f, err := os.Open(JFile)
	// if err != nil {

	// }

	return nil
}

// SaveToFile save JSON file to Cfg
func (c *DB) SaveToFile(JFile string) error {
	return nil
}

// Update changes pathCfg based on pathID, will replace
// pathDB if it's not validate
func (c *DB) Update(ID pathID, input pathDB) {

}
