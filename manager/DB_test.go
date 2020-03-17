// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.

// All rights reserved.

// Use of this source code is governed by a GPLv3-style

// license that can be found in the LICENSE file.

package manager

import (
	"fmt"
	"testing"
)

func TestDB_loadFromFile(t *testing.T) {
	db := DB{}

	db.create()

	if err := db.loadFromFile("../testdata/test1.json"); err != nil {
		return
	}

	Params := db.get(-1)
	if Params != nil {
		fmt.Printf("%v", db.get(10))
	}
}
