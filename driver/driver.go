// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

// Package driver define driver operation. all sub-cards'
// driver should be in this package
package driver

import "net"

// SlotID is sub-card's Physical SlotID Number
type SlotID int

// NameID is sub-card's NameID
type NameID string

// IP is sub-card's IP
type IP net.IP

// WorkerID is sub-card's WorkID, WorkID is uniqual in one sub
// card scope
type WorkerID int

// IsRunning is Worker's status, on/off
type IsRunning bool

// Resource is shared between path
type Resource string