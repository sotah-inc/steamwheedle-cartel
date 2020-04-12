package state

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

type busListenFunc func(onReady chan interface{}, stop chan interface{}, onStopped chan interface{})

type busListener struct {
	call      busListenFunc
	onReady   chan interface{}
	stop      chan interface{}
	onStopped chan interface{}
}

type SubjectBusListeners map[subjects.Subject]busListenFunc

func NewBusListeners(sListeners SubjectBusListeners) BusListeners {
	out := BusListeners{}
	for subj, l := range sListeners {
		out[subj] = busListener{
			call:      l,
			onStopped: make(chan interface{}),
			onReady:   make(chan interface{}),
			stop:      make(chan interface{}),
		}
	}

	return out
}

type BusListeners map[subjects.Subject]busListener

func (ls BusListeners) Listen() {
	logging.WithField("count", len(ls)).Info("Starting bus-listeners")

	for _, l := range ls {
		l.call(l.onReady, l.stop, l.onStopped)
		<-l.onReady
	}
}

func (ls BusListeners) Stop() {
	logging.WithField("count", len(ls)).Info("Stopping bus-listeners")

	for _, l := range ls {
		l.stop <- struct{}{}
		<-l.onStopped
	}
}
