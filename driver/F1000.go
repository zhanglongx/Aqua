package driver

import (
	"fmt"
	"net"

	"github.com/zhanglongx/Aqua/comm"
)

// F1000Name ..
const F1000Name = "F1000"

// F1000 ..
type F1000 struct {
	Slot int
	IP   net.IP
}

// F1000Worker ..
type F1000Worker struct {
	workerID  int
	isRunning bool
	card      *F1000
}

// Open ..
func (f *F1000) Open() ([]Worker, error) {
	return []Worker{
		&F1000Worker{
			workerID: 0,
			card:     f,
		},
		&F1000Worker{
			workerID: 1,
			card:     f,
		},
	}, nil
}

// Close ..
func (f *F1000) Close() error {
	return nil
}

// Control ..
func (w *F1000Worker) Control(c CtlCmd, arg interface{}) interface{} {
	switch c {
	case CtlCmdStart:
		if w.isRunning {
			return nil
		}
		comm.Info.Printf("F1000 Start")
		w.isRunning = true
	case CtlCmdStop:
		if !w.isRunning {
			return nil
		}
		comm.Info.Printf("F1000 Stop")
		w.isRunning = false
	case CtlCmdName:
		return fmt.Sprintf("%s_%d_%d", F1000Name, w.card.Slot, w.workerID)
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
func (w *F1000Worker) Monitor() bool {
	return true
}

// Encode ..
func (w *F1000Worker) Encode(sess *Session) error {
	comm.Info.Printf("dst: %s:%d\n", sess.IP, sess.Ports[0])
	return nil
}
