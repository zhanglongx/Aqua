// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

package manager

import (
	"fmt"

	"github.com/zhanglongx/Aqua/driver"
)

// getWorker return generic worker's info
func getWorker(w driver.Worker, data map[string]string) {

	data[STRWORKER] = fmt.Sprintf("%v", w)
}

func getPath(p *pathRow, data map[string]string) {

	data[STRPATH] = p.PathName
	data[STRRUN] = fmt.Sprintf("%v", p.IsRunning)
	// TODO: RES
}
