// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

package driver

import (
	"fmt"
	"net"
	"os/exec"

	"github.com/zhanglongx/Aqua/comm"
)

// LocalDecoderName is the sub-card's name
const LocalDecoderName string = "local_decoder"

// LocalD is the main struct for sub-card
type LocalD struct {
	// Card Slot
	Slot int

	// Card IP
	IP net.IP
}

// LocalDWorker is the main struct for sub-card's
// Worker
type LocalDWorker struct {
	workerID int

	isRunning bool

	card *LocalD

	cmd *exec.Cmd

	port [2]int
}

// Open method
func (l *LocalD) Open(s int, IP net.IP) ([]Worker, error) {
	card := LocalD{
		Slot: s,
		IP:   IP,
	}

	return []Worker{
		&LocalDWorker{
			workerID: 0,
			card:     &card,
		},
		&LocalDWorker{
			workerID: 1,
			card:     &card,
		},
	}, nil
}

// Close method
func (l *LocalD) Close() error {
	return nil
}

// Control method
func (w *LocalDWorker) Control(c CtlCmd) interface{} {
	switch c {
	case CtlCmdStart:
		if w.isRunning == true {
			return nil
		}

		url := fmt.Sprintf("rtp://localhost:%d", w.port[0])

		w.cmd = exec.Command(vlcExe, url)
		if err := w.cmd.Start(); err != nil {
			comm.Error.Printf("run vlc failed")
			return err
		}

		w.isRunning = true

	case CtlCmdStop:
		if w.isRunning == false {
			return nil
		}

		fmt.Printf("Waiting for closing VLC manually")
		if err := w.cmd.Wait(); err != nil {
			comm.Error.Printf("vlc exit with error")
			return err
		}

		w.isRunning = false

	case CtlCmdName:
		return fmt.Sprintf("%s_%d_%d", LocalDecoderName,
			w.card.Slot, w.workerID)

	case CtlCmdIP:
		return w.card.IP

	case CtlCmdSlot:
		return w.card.Slot

	case CtlCmdWorkerID:
		return w.workerID

	default:
	}
	return nil
}

// Decode method
func (w *LocalDWorker) Decode(sess *Session) error {

	w.port[0] = sess.Ports[0]
	return nil
}