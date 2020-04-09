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

// LocalEncoderName is the sub-card's name
const LocalEncoderName string = "local_encoder"

const vlcExe = "c:\\Program Files\\VideoLAN\\VLC\\vlc.exe"
const soutTpl = "#transcode{vcodec=h264,vb=300,acodec=mpga,ab=128,channels=2,samplerate=44100,scodec=none}:rtp{dst=%s,port=%d}"

// LocalE is the main struct for sub-card
type LocalE struct {
	// Card Slot
	Slot int

	// Card IP
	IP net.IP
}

// LocalEWorker is the main struct for sub-card's
// Worker
type LocalEWorker struct {
	workerID int

	isRunning bool

	card *LocalE

	cmd *exec.Cmd

	dst  net.IP
	port [2]int
}

// Open method
func (l *LocalE) Open() ([]Worker, error) {
	return []Worker{
		&LocalEWorker{
			workerID: 0,
			card:     l,
		},
		&LocalEWorker{
			workerID: 1,
			card:     l,
		},
	}, nil
}

// Close method
func (l *LocalE) Close() error {
	return nil
}

// Control method
func (w *LocalEWorker) Control(c CtlCmd) interface{} {
	switch c {
	case CtlCmdStart:
		if w.isRunning == true {
			return nil
		}

		sout := fmt.Sprintf(soutTpl, w.dst, w.port[0])

		w.cmd = exec.Command(vlcExe,
			"d:\\Streams\\D1_1M_9330.ts",
			"--sout", sout)
		if err := w.cmd.Start(); err != nil {
			comm.Error.Printf("run vlc failed")
			return err
		}

		w.isRunning = true

	case CtlCmdStop:
		if w.isRunning == false {
			return nil
		}

		fmt.Printf("Waiting for closing VLC manually\n")
		if err := w.cmd.Wait(); err != nil {
			comm.Error.Printf("vlc exit with error")
			return err
		}

		w.isRunning = false

	case CtlCmdName:
		return fmt.Sprintf("%s_%d_%d", LocalEncoderName,
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

// Encode method
func (w *LocalEWorker) Encode(sess *Session) error {

	w.dst = sess.IP
	w.port[0] = sess.Ports[0]
	return nil
}
