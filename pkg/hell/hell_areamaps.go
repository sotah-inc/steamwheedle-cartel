package hell

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/hell/state"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/gameversions"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func getAreaMapDocumentName(version gameversions.GameVersion, id int) string {
	return fmt.Sprintf("games/%s/areamaps/%d", version, id)
}

type AreaMap struct {
	State state.State `firestore:"state"`
}

func (c Client) GetAreaMap(gameVersion gameversions.GameVersion, id int) (AreaMap, error) {
	areaMapRef, err := c.FirmDocument(getAreaMapDocumentName(gameVersion, id))
	if err != nil {
		return AreaMap{}, err
	}

	docsnap, err := areaMapRef.Get(c.Context)
	if err != nil {
		return AreaMap{}, err
	}

	var areaMap AreaMap
	if err := docsnap.DataTo(&areaMap); err != nil {
		return AreaMap{}, err
	}

	return areaMap, nil
}

type FilterInNonExistJob struct {
	Id  int
	Err error
}

func (job FilterInNonExistJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"id":  job.Id,
		"err": job.Err.Error(),
	}
}

func (c Client) FilterInNonExist(gameVersion gameversions.GameVersion, ids []int) ([]int, error) {
	// spawning workers
	in := make(chan int)
	out := make(chan FilterInNonExistJob)
	worker := func() {
		for id := range in {
			docRef := c.Doc(getAreaMapDocumentName(gameVersion, id))
			_, err := docRef.Get(c.Context)
			if err == nil {
				continue
			}

			if status.Code(err) == codes.NotFound {
				out <- FilterInNonExistJob{
					Id:  id,
					Err: nil,
				}

				continue
			}

			out <- FilterInNonExistJob{
				Id:  id,
				Err: err,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(8, worker, postWork)

	// spinning it up
	go func() {
		for _, id := range ids {
			in <- id
		}

		close(in)
	}()

	// waiting for results to drain out
	var results []int
	for job := range out {
		if job.Err != nil {
			return []int{}, job.Err
		}

		results = append(results, job.Id)
	}

	return results, nil
}
