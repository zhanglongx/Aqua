// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

package driver

import (
	"fmt"
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

// Resource is resource shared between path
type Resource interface {
	URL() string
}

// SDP shared between path
type SDP struct {
	CodecVideo int
	CodecAudio int

	PtVideo int
	PtAudio int
}

// InnerRes is resource shared between inner ports
type InnerRes struct {
	// PathID will used by downstream
	PathID int

	// IP is the resource IP, downstream will use this
	// IP to fetch the stream
	IP net.IP

	// Port is the resource port
	Port []int

	// SDP
	SDP SDP
}

// OutterRes is resource to be published
type OutterRes struct {
	// URL to be published
	Rtsp string
}

// URL return url in form of card://, will be used
// by downstream
func (ir InnerRes) URL() string {
	if len(ir.Port) == 1 {
		return fmt.Sprintf("card://%v:88/%d", ir.IP, ir.Port[0])
	}

	return fmt.Sprintf("card://%v:88/%d_%d", ir.IP, ir.Port[0], ir.Port[1])
}

// URL return rtsp url
func (or OutterRes) URL() string {
	return or.Rtsp
}
