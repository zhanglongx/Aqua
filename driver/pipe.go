// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

package driver

import (
	"errors"
	"net"
)

// Video Codec
const (
	VideoH264 = iota
	VideoHEVC
)

// Audio Codec
const (
	AudioG711a = iota
	AudioG711mu
	AudioMPGA
)

// SDP shared between path
type SDP struct {
	CodecVideo int
	CodecAudio int

	PtVideo int
	PtAudio int
}

// Pipe is pipline shared between workers
type Pipe struct {

	// InPort is the data upstream ports, encoder
	// uses this as sendto port
	InPort []int

	// OutPort is the data downstream ports, decoder
	// uses this as recvfrom port
	OutPort []int

	// OutIp is the data downstream IP
	OutIP net.IP

	// SDP
	SDP SDP
}

var (
	errAllocPipeF = errors.New("Alloc Pipe failed")
)

// allocIn allocs InPort for a Pipe
func (pi *Pipe) allocIn(w Worker) {
	s := GetWorkerSlot(w)
	id := GetWorkerWorkerID(w)

	pi.InPort[0] = 8000 + 64*s + id
	pi.InPort[0] = pi.InPort[0] + 1
}

// allocOut allocs OutPort for a Pipe
func (pi *Pipe) allocOut(w Worker) {
	id := GetWorkerWorkerID(w)

	pi.InPort[0] = 8000 + id
	pi.InPort[0] = pi.InPort[0] + 1

	pi.OutIP = GetWorkerWorkerIP(w)
}
