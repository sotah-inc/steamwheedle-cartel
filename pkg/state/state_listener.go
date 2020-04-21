package state

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

type ListenStopChan chan interface{}

type listenFunc func(stop ListenStopChan) error

type listener struct {
	call     listenFunc
	stopChan ListenStopChan
}

func NewSubjectListeners(sListenersList []SubjectListeners) SubjectListeners {
	out := SubjectListeners{}
	for _, sListeners := range sListenersList {
		for lSubject, lFunc := range sListeners {
			out[lSubject] = lFunc
		}
	}

	return out
}

type SubjectListeners map[subjects.Subject]listenFunc

func NewListeners(sListeners SubjectListeners) Listeners {
	ls := Listeners{}
	for subj, l := range sListeners {
		ls[subj] = listener{l, make(ListenStopChan)}
	}

	return ls
}

type Listeners map[subjects.Subject]listener

func (ls Listeners) Listen() error {
	logging.WithField("listeners", len(ls)).Info("starting listeners")

	for _, l := range ls {
		if err := l.call(l.stopChan); err != nil {
			return err
		}
	}

	return nil
}

func (ls Listeners) Stop() {
	logging.Info("Stopping listeners")

	for _, l := range ls {
		l.stopChan <- struct{}{}
	}
}
