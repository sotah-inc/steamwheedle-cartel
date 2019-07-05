package store

import (
	"encoding/json"
	"io/ioutil"

	"cloud.google.com/go/storage"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/store/regions"
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
