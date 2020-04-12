package store

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"io/ioutil"
	"strconv"

	sotahState "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/state"

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

func (b BootBase) GetParentAreaMaps() (sotah.AreaMapMap, error) {
	obj, err := b.GetFirmObject("areatable-retail.csv", b.GetBucket())
	if err != nil {
		return sotah.AreaMapMap{}, err
	}

	objReader, err := obj.NewReader(b.client.Context)
	if err != nil {
		return sotah.AreaMapMap{}, err
	}

	result := sotah.AreaMapMap{}
	csvReader := csv.NewReader(objReader)
	for {
		// reading the next record
		record, err := csvReader.Read()
		if err == io.EOF {
			logging.Info("finished reading csv")

			break
		}
		if err != nil {
			logging.WithField("error", err.Error()).Error("failed to read from csv file")

			return sotah.AreaMapMap{}, err
		}

		// validating the record is a parent-areamap
		isParentAreaMap := record[4] == "0"
		if !isParentAreaMap {
			continue
		}

		// reading the record
		foundZoneId, err := strconv.Atoi(record[0])
		if err != nil {
			return sotah.AreaMapMap{}, err
		}

		areaName := record[2]

		normalizedAreaName, err := sotah.NormalizeString(areaName)
		if err != nil {
			return sotah.AreaMapMap{}, err
		}

		// writing it out to the result
		result[sotah.AreaMapId(foundZoneId)] = sotah.AreaMap{
			Id:             sotah.AreaMapId(foundZoneId),
			State:          sotahState.None,
			Name:           areaName,
			NormalizedName: normalizedAreaName,
		}
	}

	return result, nil
}

func (b BootBase) GetParentZoneIds() ([]sotah.AreaMapId, error) {
	obj, err := b.GetFirmObject("areatable-retail.csv", b.GetBucket())
	if err != nil {
		return []sotah.AreaMapId{}, err
	}

	objReader, err := obj.NewReader(b.client.Context)
	if err != nil {
		return []sotah.AreaMapId{}, err
	}

	found := map[sotah.AreaMapId]interface{}{}
	csvReader := csv.NewReader(objReader)
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			logging.Info("finished reading csv")

			break
		}
		if err != nil {
			logging.WithField("error", err.Error()).Error("failed to read from csv file")

			return []sotah.AreaMapId{}, err
		}

		if record[4] != "0" {
			continue
		}

		foundZoneId, err := strconv.Atoi(record[0])
		if err != nil {
			return []sotah.AreaMapId{}, err
		}

		found[sotah.AreaMapId(foundZoneId)] = struct{}{}
	}

	out := make([]sotah.AreaMapId, len(found))
	i := 0
	for foundZoneId := range found {
		out[i] = foundZoneId

		i += 1
	}

	return out, nil
}
