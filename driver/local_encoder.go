// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

package driver

import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/zhanglongx/Aqua/comm"
)

// LocalEncoderName is the sub-card's name
const LocalEncoderName NameID = "local_encoder"

// LocalE is the main struct for sub-card
type LocalE struct {
}

// LocalEWorker is the main struct for sub-card's
// Worker
type LocalEWorker struct {
	Slot SlotID

	WorkerID WorkerID

	IP IP
}

// Open method
func (l *LocalE) Open(s SlotID, IP IP) []Worker {
	var w *LocalEWorker = &LocalEWorker{
		Slot:     s,
		WorkerID: 0,
		IP:       IP,
	}

	return []Worker{w}
}

// Close method
func (l *LocalE) Close() error {
	return nil
}

// String method
func (w *LocalEWorker) String() string {
	return fmt.Sprintf("%s_%d_%d", LocalEncoderName, w.Slot, w.WorkerID)
}

// Control method
func (w *LocalEWorker) Control(c CtlCmd) interface{} {
	switch c {
	case CtlCmdStart:
		var cmd = exec.Command("ffmpeg", "-i", "d:\\Streams\\D1_1M_9330.ts",
			"-vcode", "copy", "http://localhost:1234/feed1.ffm")
		if err := cmd.Start(); err != nil {
			comm.Error.Printf("run ffmpeg failed\n")
			return errors.New("run ffmpeg failed")
		}
	case CtlCmdStop:
	default:
	}
	return nil
}

// Encode method
func (w *LocalEWorker) Encode() Resource {
	var r string
	r = fmt.Sprintf("rtsp://%v:%d/test1.sdp", w.IP, 1235)

	return Resource(r)
}
