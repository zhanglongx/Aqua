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
	HW: "ens33",

	TransitSvr: net.IPv4(10, 1, 41, 150),
	// TransitSvr: net.IPv4(192, 168, 17, 133),

	EPDir:  "testdata",
	EPFile: "encode.json",
	EPNeed: []string{"C9820Enc", "9550Av3Enc"},

	DPDir:  "testdata",
	DPFile: "decode.json",
	DPNeed: []string{"C9820Dec", "9550Av3Dec"},

	IsHTTPPipeOn: true,
}
