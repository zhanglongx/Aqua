// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

package common

import (
	"log"
	"os"
)

// Info is log[Info] output
var Info = log.New(os.Stderr, "INFO: ",
	log.Ldate|log.Ltime|log.Lshortfile)

// Warning is log[Warning] output
var Warning = log.New(os.Stderr, "WARNING: ",
	log.Ldate|log.Ltime|log.Lshortfile)

// Error is log[Error] output
var Error = log.New(os.Stderr, "ERROR: ",
	log.Ldate|log.Ltime|log.Lshortfile)
