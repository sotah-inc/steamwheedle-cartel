package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"

	"github.com/sirupsen/logrus"
	"github.com/twinj/uuid"
	dCodes "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	mCodes "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

type RequestError struct {
	Code    mCodes.Code
	Message string
}

// bus-listener functionality
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

// listener functionality
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
	logging.WithField("listeners", len(ls)).Info("Starting listeners")

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

// state
func NewState(runId uuid.UUID, useGCloud bool) State {
	return State{RunID: runId, UseGCloud: useGCloud}
}

type State struct {
	RunID        uuid.UUID
	Listeners    Listeners
	BusListeners BusListeners
	UseGCloud    bool
}

type RealmTimeTuple struct {
	Realm      sotah.Realm
	TargetTime time.Time
}

type RealmTimes map[blizzardv2.RealmSlug]RealmTimeTuple

type RegionRealmTimes map[blizzardv2.RegionName]RealmTimes

func DatabaseCodeToMessengerCode(dCode dCodes.Code) mCodes.Code {
	switch dCode {
	case dCodes.Ok:
		return mCodes.Ok
	case dCodes.Blank:
		return mCodes.Blank
	case dCodes.GenericError:
		return mCodes.GenericError
	case dCodes.MsgJSONParseError:
		return mCodes.MsgJSONParseError
	case dCodes.NotFound:
		return mCodes.NotFound
	case dCodes.UserError:
		return mCodes.UserError
	}

	return mCodes.Blank
}

type ItemBlacklist []blizzardv2.ItemId

func (ib ItemBlacklist) IsPresent(itemId blizzardv2.ItemId) bool {
	for _, blacklistItemId := range ib {
		if blacklistItemId == itemId {
			return true
		}
	}

	return false
}

func (sta State) NewRegions() (sotah.RegionList, error) {
	msg, err := func() (messenger.Message, error) {
		attempts := 0

		for {
			out, err := sta.IO.Messenger.Request(string(subjects.Boot), []byte{})
			if err == nil {
				return out, nil
			}

			attempts++

			if attempts >= 20 {
				return messenger.Message{}, fmt.Errorf("failed to fetch boot message after %d attempts", attempts)
			}

			logrus.WithField("attempt", attempts).Info("Requested boot, sleeping until next")

			time.Sleep(250 * time.Millisecond)
		}
	}()
	if err != nil {
		return sotah.RegionList{}, err
	}

	if msg.Code != mCodes.Ok {
		return nil, errors.New(msg.Err)
	}

	boot := BootResponse{}
	if err := json.Unmarshal([]byte(msg.Data), &boot); err != nil {
		return sotah.RegionList{}, err
	}

	return boot.Regions, nil
}

func NewTokenHistoryRequest(data []byte) (TokenHistoryRequest, error) {
	var out TokenHistoryRequest
	if err := json.Unmarshal(data, &out); err != nil {
		return TokenHistoryRequest{}, err
	}

	return out, nil
}

type TokenHistoryRequest struct {
	RegionName string `json:"region_name"`
}

func NewAuctionsStatsRequest(data []byte) (AuctionsStatsRequest, error) {
	var out AuctionsStatsRequest
	if err := json.Unmarshal(data, &out); err != nil {
		return AuctionsStatsRequest{}, err
	}

	return out, nil
}

type AuctionsStatsRequest struct {
	RegionName string `json:"region_name"`
	RealmSlug  string `json:"realm_slug"`
}
