package scenario

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/isucon10-qualify/isucon10-qualify/bench/asset"
	"github.com/isucon10-qualify/isucon10-qualify/bench/client"
	"github.com/isucon10-qualify/isucon10-qualify/bench/fails"
	"github.com/morikuni/failure"
)

const (
	NumOfVerifyChairDetail                = 5
	NumOfVerifyChairSearchCondition       = 1
	NumOfVerifyChairSearch                = 5
	NumOfVerifyEstateDetail               = 5
	NumOfVerifyEstateSearchCondition      = 1
	NumOfVerifyEstateSearch               = 5
	NumOfVerifyLowPricedChair             = 1
	NumOfVerifyLowPricedEstate            = 1
	NumOfVerifyRecommendedEstateWithChair = 5
	NumOfVerifyEstateNazotte              = 5
)

var (
	ignoreChairUnexported  = cmpopts.IgnoreUnexported(asset.Chair{})
	ignoreEstateUnexported = cmpopts.IgnoreUnexported(asset.Estate{})
	ignoreEstateLatitude   = cmpopts.IgnoreFields(asset.Estate{}, "Latitude")
	ignoreEstateLongitude  = cmpopts.IgnoreFields(asset.Estate{}, "Longitude")
)

type Request struct {
	Method   string `json:"method"`
	Resource string `json:"resource"`
	Query    string `json:"query"`
	Body     string `json:"body"`
}

type Response struct {
	StatusCode int    `json:"statusCode"`
	Body       string `json:"body"`
}

type Snapshot struct {
	Request  Request  `json:"request"`
	Response Response `json:"response"`
}

func loadSnapshotFromFile(filePath string) (*Snapshot, error) {
	raw, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var snapshot *Snapshot
	err = json.Unmarshal(raw, &snapshot)
	if err != nil {
		return nil, err
	}

	return snapshot, nil
}

func verifyChairDetail(ctx context.Context, c *client.Client, filePath string) error {
	snapshot, err := loadSnapshotFromFile(filePath)
	if err != nil {
		return failure.Translate(err, fails.ErrBenchmarker, failure.Message("GET /api/chair/id: Snapshot????????????????????????????????????"))
	}

	idx := strings.LastIndex(snapshot.Request.Resource, "/")
	if idx == -1 || idx == len(snapshot.Request.Resource)-1 {
		return failure.Translate(err, fails.ErrBenchmarker, failure.Message("GET /api/chair/:id: ?????????Snapshot??????"), failure.Messagef("snapshot: %s", filePath))
	}

	id := snapshot.Request.Resource[idx+1:]
	actual, err := c.GetChairDetailFromID(ctx, id)

	switch snapshot.Response.StatusCode {
	case http.StatusOK:
		if err != nil {
			return failure.Translate(err, fails.ErrApplication, failure.Message("GET /api/chair/:id: ??????????????????????????????"))
		}

		var expected *asset.Chair
		err = json.Unmarshal([]byte(snapshot.Response.Body), &expected)
		if err != nil {
			return failure.Translate(err, fails.ErrBenchmarker, failure.Message("GET /api/chair/:id: Snapshot???Response Body???Unmarshal?????????????????????????????????"), failure.Messagef("snapshot: %s", filePath))
		}

		if actual == nil {
			return failure.New(fails.ErrApplication, failure.Message("GET /api/chair/:id: ??????????????????????????????"), failure.Messagef("snapshot: %s", filePath))
		}

		if !cmp.Equal(*expected, *actual, ignoreChairUnexported) {
			return failure.New(fails.ErrApplication, failure.Message("GET /api/chair/:id: ??????????????????????????????"), failure.Messagef("snapshot: %s", filePath))
		}

	case http.StatusNotFound:
		if actual != nil {
			return failure.New(fails.ErrApplication, failure.Message("GET /api/chair/:id: ??????????????????????????????"))
		}
		if err != nil {
			return failure.Translate(err, fails.ErrApplication, failure.Message("GET /api/chair/:id: ??????????????????????????????"))
		}

	default:
		if err == nil {
			return failure.New(fails.ErrApplication, failure.Message("GET /api/chair/:id: ??????????????????????????????"))
		}
	}

	return nil
}

func verifyChairSearchCondition(ctx context.Context, c *client.Client, filePath string) error {
	snapshot, err := loadSnapshotFromFile(filePath)
	if err != nil {
		return failure.Translate(err, fails.ErrBenchmarker, failure.Message("GET /api/chair/search/condition: Snapshot????????????????????????????????????"))
	}

	actual, err := c.GetChairSearchCondition(ctx)

	switch snapshot.Response.StatusCode {
	case http.StatusOK:
		if err != nil {
			return failure.Translate(err, fails.ErrApplication, failure.Message("GET /api/chair/search/condition: ??????????????????????????????"))
		}

		var expected *asset.ChairSearchCondition
		err = json.Unmarshal([]byte(snapshot.Response.Body), &expected)
		if err != nil {
			return failure.Translate(err, fails.ErrBenchmarker, failure.Message("GET /api/chair/search/condition: Snapshot???Response Body???Unmarshal?????????????????????????????????"), failure.Messagef("snapshot: %s", filePath))
		}

		if !cmp.Equal(*expected, *actual, ignoreChairUnexported) {
			return failure.New(fails.ErrApplication, failure.Message("GET /api/chair/search/condition: ??????????????????????????????"), failure.Messagef("snapshot: %s", filePath))
		}

	default:
		if err == nil {
			return failure.New(fails.ErrApplication, failure.Message("GET /api/chair/search/condition: ??????????????????????????????"))
		}
	}

	return nil
}

func verifyChairSearch(ctx context.Context, c *client.Client, filePath string) error {
	snapshot, err := loadSnapshotFromFile(filePath)
	if err != nil {
		return failure.Translate(err, fails.ErrBenchmarker, failure.Message("GET /api/chair/search: Snapshot????????????????????????????????????"))
	}

	q, err := url.ParseQuery(snapshot.Request.Query)
	if err != nil {
		return failure.Translate(err, fails.ErrBenchmarker, failure.Message("GET /api/chair/search: Request Query???Unmarshal?????????????????????????????????"))
	}

	actual, err := c.SearchChairsWithQuery(ctx, q)

	switch snapshot.Response.StatusCode {
	case http.StatusOK:
		if err != nil {
			return failure.Translate(err, fails.ErrApplication, failure.Message("GET /api/chair/search: ??????????????????????????????"))
		}

		var expected *client.ChairsResponse
		err = json.Unmarshal([]byte(snapshot.Response.Body), &expected)
		if err != nil {
			return failure.Translate(err, fails.ErrBenchmarker, failure.Message("GET /api/chair/search: Snapshot???Response Body???Unmarshal?????????????????????????????????"), failure.Messagef("snapshot: %s", filePath))
		}

		if !cmp.Equal(*expected, *actual, ignoreChairUnexported) {
			return failure.New(fails.ErrApplication, failure.Message("GET /api/chair/search: ??????????????????????????????"), failure.Messagef("snapshot: %s", filePath))
		}

	default:
		if err == nil {
			return failure.New(fails.ErrApplication, failure.Message("GET /api/chair/search: ??????????????????????????????"))
		}
	}

	return nil
}

func verifyEstateDetail(ctx context.Context, c *client.Client, filePath string) error {
	snapshot, err := loadSnapshotFromFile(filePath)
	if err != nil {
		return failure.Translate(err, fails.ErrBenchmarker, failure.Message("GET /api/estate/id: Snapshot????????????????????????????????????"))
	}

	idx := strings.LastIndex(snapshot.Request.Resource, "/")
	if idx == -1 || idx == len(snapshot.Request.Resource)-1 {
		return failure.Translate(err, fails.ErrBenchmarker, failure.Message("GET /api/estate/:id: ?????????Snapshot??????"), failure.Messagef("snapshot: %s", filePath))
	}

	id := snapshot.Request.Resource[idx+1:]
	actual, err := c.GetEstateDetailFromID(ctx, id)

	switch snapshot.Response.StatusCode {
	case http.StatusOK:
		if err != nil {
			return failure.Translate(err, fails.ErrApplication, failure.Message("GET /api/estate/:id: ??????????????????????????????"))
		}

		var expected *asset.Estate
		err = json.Unmarshal([]byte(snapshot.Response.Body), &expected)
		if err != nil {
			return failure.Translate(err, fails.ErrBenchmarker, failure.Message("GET /api/estate/:id: Snapshot???Response Body???Unmarshal?????????????????????????????????"), failure.Messagef("snapshot: %s", filePath))
		}

		if !cmp.Equal(*expected, *actual, ignoreEstateUnexported, ignoreEstateLatitude, ignoreEstateLongitude) {
			return failure.New(fails.ErrApplication, failure.Message("GET /api/estate/:id: ??????????????????????????????"), failure.Messagef("snapshot: %s", filePath))
		}

	default:
		if err == nil {
			return failure.New(fails.ErrApplication, failure.Message("GET /api/estate/:id: ??????????????????????????????"))
		}
	}

	return nil
}

func verifyEstateSearchCondition(ctx context.Context, c *client.Client, filePath string) error {
	snapshot, err := loadSnapshotFromFile(filePath)
	if err != nil {
		return failure.Translate(err, fails.ErrBenchmarker, failure.Message("GET /api/estate/search/condition: Snapshot????????????????????????????????????"))
	}

	actual, err := c.GetEstateSearchCondition(ctx)

	switch snapshot.Response.StatusCode {
	case http.StatusOK:
		if err != nil {
			return failure.Translate(err, fails.ErrApplication, failure.Message("GET /api/estate/search/condition: ??????????????????????????????"))
		}

		var expected *asset.EstateSearchCondition
		err = json.Unmarshal([]byte(snapshot.Response.Body), &expected)
		if err != nil {
			return failure.Translate(err, fails.ErrBenchmarker, failure.Message("GET /api/estate/search/condition: Snapshot???Response Body???Unmarshal?????????????????????????????????"), failure.Messagef("snapshot: %s", filePath))
		}

		if !cmp.Equal(*expected, *actual) {
			return failure.New(fails.ErrApplication, failure.Message("GET /api/estate/search/condition: ??????????????????????????????"), failure.Messagef("snapshot: %s", filePath))
		}

	default:
		if err == nil {
			return failure.New(fails.ErrApplication, failure.Message("GET /api/estate/search/condition: ??????????????????????????????"))
		}
	}

	return nil
}

func verifyEstateSearch(ctx context.Context, c *client.Client, filePath string) error {
	snapshot, err := loadSnapshotFromFile(filePath)
	if err != nil {
		return failure.Translate(err, fails.ErrBenchmarker, failure.Message("GET /api/estate/search: Snapshot????????????????????????????????????"))
	}

	q, err := url.ParseQuery(snapshot.Request.Query)
	if err != nil {
		return failure.Translate(err, fails.ErrBenchmarker, failure.Message("GET /api/estate/search: Request Query???Unmarshal?????????????????????????????????"))
	}

	actual, err := c.SearchEstatesWithQuery(ctx, q)

	switch snapshot.Response.StatusCode {
	case http.StatusOK:
		if err != nil {
			return failure.Translate(err, fails.ErrApplication, failure.Message("GET /api/estate/search: ??????????????????????????????"))
		}

		var expected *client.EstatesResponse
		err = json.Unmarshal([]byte(snapshot.Response.Body), &expected)
		if err != nil {
			return failure.Translate(err, fails.ErrBenchmarker, failure.Message("GET /api/estate/search: Snapshot???Response Body???Unmarshal?????????????????????????????????"), failure.Messagef("snapshot: %s", filePath))
		}

		if !cmp.Equal(*expected, *actual, ignoreEstateUnexported, ignoreEstateLatitude, ignoreEstateLongitude) {
			return failure.New(fails.ErrApplication, failure.Message("GET /api/estate/search: ??????????????????????????????"), failure.Messagef("snapshot: %s", filePath))
		}

	default:
		if err == nil {
			return failure.New(fails.ErrApplication, failure.Message("GET /api/estate/search: ??????????????????????????????"))
		}
	}

	return nil
}

func verifyLowPricedChair(ctx context.Context, c *client.Client, filePath string) error {
	snapshot, err := loadSnapshotFromFile(filePath)
	if err != nil {
		return failure.Translate(err, fails.ErrBenchmarker, failure.Message("GET /api/chair/low_priced: Snapshot????????????????????????????????????"))
	}

	actual, err := c.GetLowPricedChair(ctx)

	switch snapshot.Response.StatusCode {
	case http.StatusOK:
		if err != nil {
			return failure.Translate(err, fails.ErrApplication, failure.Message("GET /api/chair/low_priced: ??????????????????????????????"))
		}

		var expected *client.ChairsResponse
		err = json.Unmarshal([]byte(snapshot.Response.Body), &expected)
		if err != nil {
			return failure.Translate(err, fails.ErrBenchmarker, failure.Message("GET /api/chair/low_priced: Snapshot???Response Body???Unmarshal?????????????????????????????????"), failure.Messagef("snapshot: %s", filePath))
		}

		if !cmp.Equal(*expected, *actual, ignoreChairUnexported) {
			return failure.New(fails.ErrApplication, failure.Message("GET /api/chair/low_priced: ??????????????????????????????"), failure.Messagef("snapshot: %s", filePath))
		}

	default:
		if err == nil {
			return failure.New(fails.ErrApplication, failure.Message("GET /api/chair/low_priced: ??????????????????????????????"))
		}
	}

	return nil
}

func verifyLowPricedEstate(ctx context.Context, c *client.Client, filePath string) error {
	snapshot, err := loadSnapshotFromFile(filePath)
	if err != nil {
		return failure.Translate(err, fails.ErrBenchmarker, failure.Message("GET /api/estate/low_priced: Snapshot????????????????????????????????????"))
	}

	actual, err := c.GetLowPricedEstate(ctx)

	switch snapshot.Response.StatusCode {
	case http.StatusOK:
		if err != nil {
			return failure.Translate(err, fails.ErrApplication, failure.Message("GET /api/estate/low_priced: ??????????????????????????????"))
		}

		var expected *client.EstatesResponse
		err = json.Unmarshal([]byte(snapshot.Response.Body), &expected)
		if err != nil {
			return failure.Translate(err, fails.ErrBenchmarker, failure.Message("GET /api/estate/low_priced: Snapshot???Response Body???Unmarshal?????????????????????????????????"), failure.Messagef("snapshot: %s", filePath))
		}

		if !cmp.Equal(*expected, *actual, ignoreEstateUnexported, ignoreEstateLatitude, ignoreEstateLongitude) {
			return failure.New(fails.ErrApplication, failure.Message("GET /api/estate/low_priced: ??????????????????????????????"), failure.Messagef("snapshot: %s", filePath))
		}

	default:
		if err == nil {
			return failure.New(fails.ErrApplication, failure.Message("GET /api/estate/low_priced: ??????????????????????????????"))
		}
	}

	return nil
}

func verifyRecommendedEstateWithChair(ctx context.Context, c *client.Client, filePath string) error {
	snapshot, err := loadSnapshotFromFile(filePath)
	if err != nil {
		return failure.Translate(err, fails.ErrBenchmarker, failure.Message("GET /api/recommended_estate/:id: Snapshot????????????????????????????????????"))
	}

	idx := strings.LastIndex(snapshot.Request.Resource, "/")
	if idx == -1 || idx == len(snapshot.Request.Resource)-1 {
		return failure.Translate(err, fails.ErrBenchmarker, failure.Message("GET /api/recommended_estate/:id: ?????????Snapshot??????"), failure.Messagef("snapshot: %s", filePath))
	}
	id, err := strconv.ParseInt(snapshot.Request.Resource[idx+1:], 10, 64)
	if err != nil {
		return failure.Translate(err, fails.ErrBenchmarker, failure.Message("GET /api/recommended_estate/:id: ?????????Snapshot??????"), failure.Messagef("snapshot: %s", filePath))
	}

	actual, err := c.GetRecommendedEstatesFromChair(ctx, id)

	switch snapshot.Response.StatusCode {
	case http.StatusOK:
		if err != nil {
			return failure.Translate(err, fails.ErrApplication, failure.Message("GET /api/recommended_estate/:id: ??????????????????????????????"))
		}

		var expected *client.EstatesResponse
		err = json.Unmarshal([]byte(snapshot.Response.Body), &expected)
		if err != nil {
			return failure.Translate(err, fails.ErrBenchmarker, failure.Message("GET /api/recommended_estate/:id: Snapshot???Response Body???Unmarshal?????????????????????????????????"), failure.Messagef("snapshot: %s", filePath))
		}
		if !cmp.Equal(*expected, *actual, ignoreEstateUnexported, ignoreEstateLatitude, ignoreEstateLongitude) {
			return failure.New(fails.ErrApplication, failure.Message("GET /api/recommended_estate/:id: ??????????????????????????????"), failure.Messagef("snapshot: %s", filePath))
		}

	default:
		if err == nil {
			return failure.New(fails.ErrApplication, failure.Message("GET /api/recommended_estate/:id: ??????????????????????????????"))
		}
	}

	return nil
}

func verifyEstateNazotte(ctx context.Context, c *client.Client, filePath string) error {
	snapshot, err := loadSnapshotFromFile(filePath)
	if err != nil {
		return failure.Translate(err, fails.ErrBenchmarker, failure.Message("POST /api/estate/nazotte: Snapshot????????????????????????????????????"))
	}

	var coordinates *client.Coordinates
	err = json.Unmarshal([]byte(snapshot.Request.Body), &coordinates)
	if err != nil {
		return failure.Translate(err, fails.ErrBenchmarker, failure.Message("POST /api/estate/nazotte: Request Body???Unmarshal?????????????????????????????????"))
	}

	actual, err := c.SearchEstatesNazotte(ctx, coordinates)

	switch snapshot.Response.StatusCode {
	case http.StatusOK:
		if err != nil {
			return failure.Translate(err, fails.ErrApplication, failure.Message("POST /api/estate/nazotte: ??????????????????????????????"))
		}

		var expected *client.EstatesResponse
		err = json.Unmarshal([]byte(snapshot.Response.Body), &expected)
		if err != nil {
			return failure.Translate(err, fails.ErrBenchmarker, failure.Message("POST /api/estate/nazotte: Snapshot???Response Body???Unmarshal?????????????????????????????????"), failure.Messagef("snapshot: %s", filePath))
		}

		if !cmp.Equal(*expected, *actual, ignoreEstateUnexported, ignoreEstateLatitude, ignoreEstateLongitude) {
			return failure.New(fails.ErrApplication, failure.Message("POST /api/estate/nazotte: ??????????????????????????????"), failure.Messagef("snapshot: %s", filePath))
		}

	default:
		if err == nil {
			return failure.New(fails.ErrApplication, failure.Message("POST /api/estate/nazotte: ??????????????????????????????"))
		}
	}

	return nil
}

func verifyWithSnapshot(ctx context.Context, c *client.Client, snapshotsParentsDirPath string) {
	wg := sync.WaitGroup{}

	snapshotsDirPath := filepath.Join(snapshotsParentsDirPath, "chair_detail")
	snapshots, err := ioutil.ReadDir(snapshotsDirPath)
	if err != nil {
		err := failure.Translate(err, fails.ErrBenchmarker, failure.Message("GET /api/chair/:id: Snapshot????????????????????????????????????"))
		fails.Add(err)
	} else {
		for i := 0; i < NumOfVerifyChairDetail; i++ {
			wg.Add(1)
			r := rand.Intn(len(snapshots))
			go func(filePath string) {
				err := verifyChairDetail(ctx, c, filePath)
				if err != nil {
					fails.Add(err)
				}
				wg.Done()
			}(path.Join(snapshotsDirPath, snapshots[r].Name()))
		}
	}

	snapshotsDirPath = filepath.Join(snapshotsParentsDirPath, "chair_search_condition")
	snapshots, err = ioutil.ReadDir(snapshotsDirPath)
	if err != nil {
		err := failure.Translate(err, fails.ErrBenchmarker, failure.Message("GET /api/chair/search/condition: Snapshot????????????????????????????????????"))
		fails.Add(err)
	} else {
		for i := 0; i < NumOfVerifyChairSearchCondition; i++ {
			wg.Add(1)
			r := rand.Intn(len(snapshots))
			go func(filePath string) {
				err := verifyChairSearchCondition(ctx, c, filePath)
				if err != nil {
					fails.Add(err)
				}
				wg.Done()
			}(path.Join(snapshotsDirPath, snapshots[r].Name()))
		}
	}

	snapshotsDirPath = filepath.Join(snapshotsParentsDirPath, "chair_search")
	snapshots, err = ioutil.ReadDir(snapshotsDirPath)
	if err != nil {
		err := failure.Translate(err, fails.ErrBenchmarker, failure.Message("GET /api/chair/search: Snapshot????????????????????????????????????"))
		fails.Add(err)
	} else {
		for i := 0; i < NumOfVerifyChairSearch; i++ {
			wg.Add(1)
			r := rand.Intn(len(snapshots))
			go func(filePath string) {
				err := verifyChairSearch(ctx, c, filePath)
				if err != nil {
					fails.Add(err)
				}
				wg.Done()
			}(path.Join(snapshotsDirPath, snapshots[r].Name()))
		}
	}

	snapshotsDirPath = filepath.Join(snapshotsParentsDirPath, "estate_detail")
	snapshots, err = ioutil.ReadDir(snapshotsDirPath)
	if err != nil {
		err := failure.Translate(err, fails.ErrBenchmarker, failure.Message("GET /api/estate/:id: Snapshot????????????????????????????????????"))
		fails.Add(err)
	} else {
		for i := 0; i < NumOfVerifyEstateDetail; i++ {
			wg.Add(1)
			r := rand.Intn(len(snapshots))
			go func(filePath string) {
				err := verifyEstateDetail(ctx, c, filePath)
				if err != nil {
					fails.Add(err)
				}
				wg.Done()
			}(path.Join(snapshotsDirPath, snapshots[r].Name()))
		}
	}

	snapshotsDirPath = filepath.Join(snapshotsParentsDirPath, "estate_search_condition")
	snapshots, err = ioutil.ReadDir(snapshotsDirPath)
	if err != nil {
		err := failure.Translate(err, fails.ErrBenchmarker, failure.Message("GET /api/estate/search/condition: Snapshot????????????????????????????????????"))
		fails.Add(err)
	} else {
		for i := 0; i < NumOfVerifyEstateSearchCondition; i++ {
			wg.Add(1)
			r := rand.Intn(len(snapshots))
			go func(filePath string) {
				err := verifyEstateSearchCondition(ctx, c, filePath)
				if err != nil {
					fails.Add(err)
				}
				wg.Done()
			}(path.Join(snapshotsDirPath, snapshots[r].Name()))
		}
	}

	snapshotsDirPath = filepath.Join(snapshotsParentsDirPath, "estate_search")
	snapshots, err = ioutil.ReadDir(snapshotsDirPath)
	if err != nil {
		err := failure.Translate(err, fails.ErrBenchmarker, failure.Message("GET /api/estate/search: Snapshot????????????????????????????????????"))
		fails.Add(err)
	} else {
		for i := 0; i < NumOfVerifyEstateSearch; i++ {
			wg.Add(1)
			r := rand.Intn(len(snapshots))
			go func(filePath string) {
				err := verifyEstateSearch(ctx, c, filePath)
				if err != nil {
					fails.Add(err)
				}
				wg.Done()
			}(path.Join(snapshotsDirPath, snapshots[r].Name()))
		}
	}

	snapshotsDirPath = filepath.Join(snapshotsParentsDirPath, "chair_low_priced")
	snapshots, err = ioutil.ReadDir(snapshotsDirPath)
	if err != nil {
		err := failure.Translate(err, fails.ErrBenchmarker, failure.Message("GET /api/chair/low_priced: Snapshot????????????????????????????????????"))
		fails.Add(err)
	} else {
		for i := 0; i < NumOfVerifyLowPricedChair; i++ {
			wg.Add(1)
			r := rand.Intn(len(snapshots))
			go func(filePath string) {
				err := verifyLowPricedChair(ctx, c, filePath)
				if err != nil {
					fails.Add(err)
				}
				wg.Done()
			}(path.Join(snapshotsDirPath, snapshots[r].Name()))
		}
	}

	snapshotsDirPath = filepath.Join(snapshotsParentsDirPath, "estate_low_priced")
	snapshots, err = ioutil.ReadDir(snapshotsDirPath)
	if err != nil {
		err := failure.Translate(err, fails.ErrBenchmarker, failure.Message("GET /api/estate/low_priced: Snapshot????????????????????????????????????"))
		fails.Add(err)
	} else {
		for i := 0; i < NumOfVerifyLowPricedEstate; i++ {
			wg.Add(1)
			r := rand.Intn(len(snapshots))
			go func(filePath string) {
				err := verifyLowPricedEstate(ctx, c, filePath)
				if err != nil {
					fails.Add(err)
				}
				wg.Done()
			}(path.Join(snapshotsDirPath, snapshots[r].Name()))
		}
	}

	snapshotsDirPath = filepath.Join(snapshotsParentsDirPath, "recommended_estate_with_chair")
	snapshots, err = ioutil.ReadDir(snapshotsDirPath)
	if err != nil {
		err := failure.Translate(err, fails.ErrBenchmarker, failure.Message("GET /api/recommended_estate/:id: Snapshot????????????????????????????????????"))
		fails.Add(err)
	} else {
		for i := 0; i < NumOfVerifyRecommendedEstateWithChair; i++ {
			wg.Add(1)
			r := rand.Intn(len(snapshots))
			go func(filePath string) {
				err := verifyRecommendedEstateWithChair(ctx, c, filePath)
				if err != nil {
					fails.Add(err)
				}
				wg.Done()
			}(path.Join(snapshotsDirPath, snapshots[r].Name()))
		}
	}

	snapshotsDirPath = filepath.Join(snapshotsParentsDirPath, "estate_nazotte")
	snapshots, err = ioutil.ReadDir(snapshotsDirPath)
	if err != nil {
		err := failure.Translate(err, fails.ErrBenchmarker, failure.Message("POST /api/estate/nazotte: Snapshot????????????????????????????????????"))
		fails.Add(err)
	} else {
		for i := 0; i < NumOfVerifyEstateNazotte; i++ {
			wg.Add(1)
			r := rand.Intn(len(snapshots))
			go func(filePath string) {
				err := verifyEstateNazotte(ctx, c, filePath)
				if err != nil {
					fails.Add(err)
				}
				wg.Done()
			}(path.Join(snapshotsDirPath, snapshots[r].Name()))
		}
	}

	wg.Wait()
}
