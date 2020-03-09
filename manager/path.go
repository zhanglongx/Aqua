// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

package manager

// getPathName
func getPathName(p *pathRow) string {
	return p.PathName
}

// getPathRunning
func getPathRunning(p *pathRow) string {
	if p.IsRunning {
		return "true"
	}

	return "false"
}
