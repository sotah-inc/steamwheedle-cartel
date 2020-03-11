package store

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"io/ioutil"
	"strconv"

	"github.com/sirupsen/logrus"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"

	"cloud.google.com/go/storage"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/store/regions"
)

func NewBootBase(c Client, location regions.Region) BootBase {
	return BootBase{base{client: c, location: location}}
}

type BootBase struct {
	base
}

func (b BootBase) getBucketName() string {
	return "sotah-boot"
}

func (b BootBase) GetBucket() *storage.BucketHandle {
	return b.base.getBucket(b.getBucketName())
}

func (b BootBase) GetFirmBucket() (*storage.BucketHandle, error) {
	return b.base.getFirmBucket(b.getBucketName())
}

func (b BootBase) GetObject(name string, bkt *storage.BucketHandle) *storage.ObjectHandle {
	return b.base.getObject(name, bkt)
}

func (b BootBase) GetFirmObject(name string, bkt *storage.BucketHandle) (*storage.ObjectHandle, error) {
	return b.base.getFirmObject(name, bkt)
}

func (b BootBase) GetRegions(bkt *storage.BucketHandle) (sotah.RegionList, error) {
	regionObj, err := b.getFirmObject("regions.json.gz", bkt)
	if err != nil {
		return sotah.RegionList{}, err
	}

	reader, err := regionObj.NewReader(b.client.Context)
	if err != nil {
		return sotah.RegionList{}, err
	}

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return sotah.RegionList{}, err
	}

	var out sotah.RegionList
	if err := json.Unmarshal(data, &out); err != nil {
		return sotah.RegionList{}, err
	}

	return out, nil
}

func (b BootBase) GetBlizzardCredentials(bkt *storage.BucketHandle) (sotah.BlizzardCredentials, error) {
	credentialsObj, err := b.getFirmObject("blizzard-credentials.json", bkt)
	if err != nil {
		return sotah.BlizzardCredentials{}, err
	}

	reader, err := credentialsObj.NewReader(b.client.Context)
	if err != nil {
		return sotah.BlizzardCredentials{}, err
	}

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return sotah.BlizzardCredentials{}, err
	}

	var out sotah.BlizzardCredentials
	if err := json.Unmarshal(data, &out); err != nil {
		return sotah.BlizzardCredentials{}, err
	}

	return out, nil
}

func (b BootBase) Guard(objName string, contents string, bkt *storage.BucketHandle) (bool, error) {
	obj, err := b.GetFirmObject(objName, bkt)
	if err != nil {
		return false, err
	}
	reader, err := obj.NewReader(b.client.Context)
	if err != nil {
		return false, err
	}
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return false, err
	}

	return string(data) == contents, nil
}

func (b BootBase) GetParentZoneIds() ([]int, error) {
	obj, err := b.GetFirmObject("areatable-retail.csv", b.GetBucket())
	if err != nil {
		return []int{}, err
	}

	objReader, err := obj.NewReader(b.client.Context)
	if err != nil {
		return []int{}, err
	}

	found := map[int]interface{}{}
	csvReader := csv.NewReader(objReader)
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			logging.Info("Finished reading csv")

			break
		}
		if err != nil {
			logging.WithField("error", err.Error()).Error("failed to read from csv file")

			return []int{}, err
		}

		logging.WithFields(logrus.Fields{
			"record":         record,
			"record[4]":      record[4],
			"record[4] != 0": record[4] != "0",
		}).Info("using record")

		if record[4] != "0" {
			continue
		}

		parentZoneId, err := strconv.Atoi(record[4])
		if err != nil {
			return []int{}, err
		}

		found[parentZoneId] = struct{}{}
	}

	out := make([]int, len(found))
	i := 0
	for parentZoneId := range found {
		out[i] = parentZoneId

		i += 1
	}

	return out, nil
}
