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
const sout = "#transcode{vcodec=h264,acodec=mpga,ab=128,channels=2,samplerate=44100,scodec=none}:rtp{sdp=rtsp://:8554/test}"

// LocalE is the main struct for sub-card
type LocalE struct {
}

// LocalEWorker is the main struct for sub-card's
// Worker
type LocalEWorker struct {
	Slot int

	WorkerID int

	IP net.IP

	IsRunning bool

	cmd *exec.Cmd
}

// Open method
func (l *LocalE) Open(s int, IP net.IP) []Worker {
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

// Control method
func (w *LocalEWorker) Control(c CtlCmd) interface{} {
	switch c {
	case CtlCmdStart:
		if w.IsRunning == true {
			return nil
		}

		w.cmd = exec.Command(vlcExe,
			"d:\\Streams\\D1_1M_9330.ts",
			"--sout", sout)
		if err := w.cmd.Start(); err != nil {
			comm.Error.Printf("run vlc failed")
			return err
		}

		w.IsRunning = true

	case CtlCmdStop:
		if w.IsRunning == false {
			return nil
		}

		fmt.Printf("Waiting for closing VLC manually")
		if err := w.cmd.Wait(); err != nil {
			comm.Error.Printf("vlc exit with error")
			return err
		}

		w.IsRunning = false

	case CtlCmdName:
		return fmt.Sprintf("%s_%d_%d", LocalEncoderName,
			w.Slot, w.WorkerID)

	default:
	}
	return nil
}

// Encode method
func (w *LocalEWorker) Encode() InnerRes {

	var sdp SDP = SDP{CodecVideo: VideoH264,
		CodecAudio: AudioG711a,
		PtVideo:    96,
		PtAudio:    8,
	}

	return InnerRes{Port: []int{1235}, SDP: sdp}
}
