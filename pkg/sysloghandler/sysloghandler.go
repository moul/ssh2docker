package sysloghandler

import (
	"errors"
	"log/syslog"
	"sync"

	"github.com/apex/log"
)

type Handler struct {
	mu sync.Mutex
	w  *syslog.Writer
}

func New(network, raddr string, priority syslog.Priority, tag string) *Handler {
	w, err := syslog.Dial(network, raddr, priority, tag)
	if err != nil {
		log.Fatalf("%v", err)
	}
	return &Handler{
		w: w,
	}
}

func (h *Handler) HandleLog(e *log.Entry) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	switch e.Level {
	case log.DebugLevel:
		return h.w.Debug(e.Message)
	case log.InfoLevel:
		return h.w.Info(e.Message)
	case log.WarnLevel:
		return h.w.Warning(e.Message)
	case log.ErrorLevel:
		return h.w.Err(e.Message)
	case log.FatalLevel:
		return h.w.Crit(e.Message)
	}
	return errors.New("invalid level")
}
