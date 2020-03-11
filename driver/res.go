// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

package driver

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

// InnerRes is resource shared between inner ports
type InnerRes struct {

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
