package driver

import (
	"fmt"
	"net"

	"github.com/zhanglongx/Aqua/comm"
)

// F2000Name ..
const F2000Name = "F2000"

// F2000 ..
type F2000 struct {
	Slot int
	IP   net.IP
}

// F2000Worker ..
type F2000Worker struct {
	workerID  int
	isRunning bool
	card      *F2000
}

// Open ..
func (f *F2000) Open() ([]Worker, error) {
	return []Worker{
		&F2000Worker{
			workerID: 0,
			card:     f,
		},
		&F2000Worker{
			workerID: 1,
			card:     f,
		},
	}, nil
}

// Close ..
func (f *F2000) Close() error {
	return nil
}

// Control ..
func (w *F2000Worker) Control(c CtlCmd, arg interface{}) interface{} {
	switch c {
	case CtlCmdStart:
		if w.isRunning {
			return nil
		}
		comm.Info.Printf("F2000 Start")
		w.isRunning = true
	case CtlCmdStop:
		if !w.isRunning {
			return nil
		}
		comm.Info.Printf("F2000 Stop")
		w.isRunning = false
	case CtlCmdName:
		return fmt.Sprintf("%s_%d_%d", F2000Name, w.card.Slot, w.workerID)
	case CtlCmdIP:
		return w.card.IP
	case CtlCmdWorkerID:
		return w.workerID
	case CtlCmdSetting:
		comm.Info.Printf("Settings")
		return nil
	}
	return nil
}

// Monitor .
func (w *F2000Worker) Monitor() bool {
	return true
}

// Decode ..
func (w *F2000Worker) Decode(sess *Session) error {
	comm.Info.Printf("dst: %s:%d\n", sess.IP, sess.Ports[0])
	return nil
}
