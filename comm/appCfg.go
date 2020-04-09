// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

package comm

import "net"

// AppCfg is the global configurations of Aqua
var AppCfg = struct {
	HW string

	TransitSvr net.IP

	EPDir  string
	EPFile string
	EPNeed []string

	DPDir  string
	DPFile string
	DPNeed []string

	IsHTTPPipeOn bool
}{
	HW: "以太网",

	TransitSvr: net.IPv4(10, 1, 41, 152),

	EPDir:  "testdata",
	EPFile: "encode.json",
	EPNeed: []string{"local_encoder"},

	DPDir:  "testdata",
	DPFile: "decode.json",
	DPNeed: []string{"local_decoder"},

	IsHTTPPipeOn: true,
}
