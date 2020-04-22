// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

package driver

import (
	"fmt"
	"net"

	"github.com/zhanglongx/Aqua/comm"
)

// Dummy is a good start to write sub-card's driver.
// There are two main struct in one driver:
//
// type Dummy struct {
//    ...
// }
//
// type DummyWorker struct {
//    ...
// }
//
// Dummy is responsible for initializing sub-card,
// and un-initializing sub-card. Dummy.Open() return
// slice of DummyWorker to the manager.
//
// DummyWorker provide necessary interface{} to the
// manager. manager call the interface{} to do all
// operation.

// DummyName is the sub-card's name
const DummyName string = "Dummy"

// Dummy is the main struct for sub-card
type Dummy struct {
	Slot int

	IP net.IP
}

// DummyWorker is the main struct for sub-card's Worker
type DummyWorker struct {
	// SlotID here
	Slot int

	// WorkerID here
	WorkerID int

	// IP here
	IP net.IP
}

// Open sub-card, do initialization. And return slice of
// Worker interface{}. Here you can setup net connection
// to sub-card, and perform necessary communication with
// it, as querying the hardware version or working path
// in sub-card
func (d *Dummy) Open() ([]Worker, error) {
	var w *DummyWorker = &DummyWorker{
		Slot:     d.Slot,
		WorkerID: 0,
		IP:       d.IP,
	}

	comm.Info.Printf("Open %s successfully", DummyName)
	return []Worker{w}, nil
}

// Close sub-card, do un-initialization. a close of
// connection usually required. But you can do more
// here
func (d *Dummy) Close() error {
	comm.Info.Printf("Close %s successfully", DummyName)
	return nil
}

// Control do quering and setting, like querying version,
// setting paramenters. Return nil if ont all CtlCmd is
// supported
func (w *DummyWorker) Control(c CtlCmd, arg interface{}) interface{} {
	switch c {
	case CtlCmdName:
		return fmt.Sprintf("%s_%d_%d", LocalEncoderName, w.Slot, w.WorkerID)
	}
	return nil
}

// Report do reporting
func (w *DummyWorker) Report() []string {
	return nil
}
