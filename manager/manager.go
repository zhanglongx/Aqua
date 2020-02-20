// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

// Package manager is the core module in Aqua, deals with
// config loading, sub-card's worker setting
package manager

import (
	"errors"
)

// pathID is the path's ID
type pathID string

// InValidPathID is ID of Invalidate
const InValidPathID pathID = ""

// Manager is main struct for mananger operation
type Manager struct {
	// DB store config settings
	DB DB
}

// Init deals with initialization, should be run only once
func (m *Manager) Init(DBFile string) error {
	if err := m.DB.loadFromFile(DBFile); err != nil {
		return errors.New("load from file failed")
	}

	return nil
}

// Run start the main loop
func (m *Manager) Run() {

}
