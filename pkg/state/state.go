package state

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/blizzard"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/bus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/database"
	dCodes "github.com/sotah-inc/steamwheedle-cartel/pkg/database/codes"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/diskstore"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/hell"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/messenger"
	mCodes "github.com/sotah-inc/steamwheedle-cartel/pkg/messenger/codes"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/metric"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/resolver"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state/subjects"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/store"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/util"
	"github.com/twinj/uuid"
)

type RequestError struct {
	Code    mCodes.Code
	Message string
}

// databases
type Databases struct {
	PricelistHistoryDatabases database.PricelistHistoryDatabases
	LiveAuctionsDatabases     database.LiveAuctionsDatabases
	ItemsDatabase             database.ItemsDatabase
	MetaDatabase              database.MetaDatabase
	PubsubTopicsDatabase      database.PubsubTopicsDatabase
}

// io bundle
type IO struct {
	Resolver    resolver.Resolver
	Databases   Databases
	Messenger   messenger.Messenger
	StoreClient store.Client
	DiskStore   diskstore.DiskStore
	Reporter    metric.Reporter
	BusClient   bus.Client
	HellClient  hell.Client
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

	IO IO
}

type RealmTimeTuple struct {
	Realm      sotah.Realm
	TargetTime time.Time
}

type RealmTimes map[blizzard.RealmSlug]RealmTimeTuple

type RegionRealmTimes map[blizzard.RegionName]RealmTimes

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

func NewRealmModificationDatesRequest(data []byte) (RealmModificationDatesRequest, error) {
	var r RealmModificationDatesRequest
	if err := json.Unmarshal(data, &r); err != nil {
		return RealmModificationDatesRequest{}, err
	}

	return r, nil
}

type RealmModificationDatesRequest struct {
	RegionName string `json:"region_name"`
	RealmSlug  string `json:"realm_slug"`
}

type RealmModificationDatesResponse struct {
	sotah.RealmModificationDates
}

func (r RealmModificationDatesResponse) EncodeForDelivery() ([]byte, error) {
	return json.Marshal(r)
}

type ItemBlacklist []blizzard.ItemID

func (ib ItemBlacklist) IsPresent(itemId blizzard.ItemID) bool {
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

type BootResponse struct {
	Regions     sotah.RegionList     `json:"regions"`
	ItemClasses blizzard.ItemClasses `json:"item_classes"`
	Expansions  []sotah.Expansion    `json:"expansions"`
	Professions []sotah.Profession   `json:"professions"`
}

type SessionSecretData struct {
	SessionSecret string `json:"session_secret"`
}

func NewItemsRequest(payload []byte) (ItemsRequest, error) {
	iRequest := &ItemsRequest{}
	err := json.Unmarshal(payload, &iRequest)
	if err != nil {
		return ItemsRequest{}, err
	}

	return *iRequest, nil
}

type ItemsRequest struct {
	ItemIds []blizzard.ItemID `json:"itemIds"`
}

func (iRequest ItemsRequest) Resolve(sta State) (sotah.ItemsMap, error) {
	return sta.IO.Databases.ItemsDatabase.FindItems(iRequest.ItemIds)
}

type ItemsResponse struct {
	Items sotah.ItemsMap `json:"items"`
}

func (iResponse ItemsResponse) EncodeForMessage() (string, error) {
	encodedResult, err := json.Marshal(iResponse)
	if err != nil {
		return "", err
	}

	gzippedResult, err := util.GzipEncode(encodedResult)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(gzippedResult), nil
}
