package test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"device-analytics/entities"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//unique devices accepts timestamps of YYYYMMDD and YYYYMMDDHH
//does not require full month in range: 20210101 to 20210102 returns monthly data

func testURL(suffix string) string {
	var fallback = "http://localhost:8080/metrics/unique-devices"
	var res string
	if res = os.Getenv("API_URL"); res == "" {
		return fmt.Sprintf("%s/%s", fallback, suffix)
	}
	return fmt.Sprintf("%s/%s", strings.TrimRight(res, "/"), suffix)
}

func runQuery(t *testing.T, project string, sites string, granularity string, start int, end int) entities.UniqueDevicesResponse {

	res, err := http.Get(testURL(fmt.Sprintf("%s/%s/%s/%s/%s", project, sites, granularity, strconv.FormatInt(int64(start), 10), strconv.FormatInt(int64(end), 10))))

	require.NoError(t, err, "Invalid http request")

	require.Equal(t, http.StatusOK, res.StatusCode, "Wrong status code")

	body, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err, "Unable to read response")

	n := entities.UniqueDevicesResponse{}
	err = json.Unmarshal(body, &n)

	require.NoError(t, err, "Unable to unmarshal response body")
	return n
}

func TestUniqueDevices(t *testing.T) {
	t.Run("should return 404 for an invalid route", func(t *testing.T) {
		// Leading slash should result in an invalid url
		res, err := http.Get(testURL("/en.wikipedia/all-sites/daily/20210101/20210201"))

		require.NoError(t, err, "Invalid http request")

		require.Equal(t, http.StatusNotFound, res.StatusCode, "Wrong status code")
	})

	t.Run("should return 200 for expected parameters", func(t *testing.T) {

		res, err := http.Get(testURL("en.wikipedia.org/all-sites/daily/20210101/20210201"))

		require.NoError(t, err, "Invalid http request")

		require.Equal(t, http.StatusOK, res.StatusCode, "Wrong status code")
	})

	t.Run("should return 400 when parameters are wrong", func(t *testing.T) {

		res, err := http.Get(testURL("wrong-project/wrong-sites/wrong-granularity/00000000/00000000"))

		require.NoError(t, err, "Invalid http request")

		require.Equal(t, http.StatusBadRequest, res.StatusCode, "Wrong status code")
	})

	t.Run("should return 400 when start is after end", func(t *testing.T) {

		res, err := http.Get(testURL("en.wikipedia/all-sites/daily/20210201/20210101"))

		require.NoError(t, err, "Invalid http request")

		require.Equal(t, http.StatusBadRequest, res.StatusCode, "Wrong status code")

	})

	t.Run("should return 400 when timestamp is invalid", func(t *testing.T) {

		res, err := http.Get(testURL("en.wikipedia/all-sites/daily/0000/0000"))

		require.NoError(t, err, "Invalid http request")

		require.Equal(t, http.StatusBadRequest, res.StatusCode, "Wrong status code")

	})

	t.Run("should return 404 for invalid route", func(t *testing.T) {

		res, err := http.Get(testURL("en.wikipedia.org/wiki/.invalid/all-sites/daily/20190529/20200229"))

		require.NoError(t, err, "Invalid http request")

		require.Equal(t, http.StatusNotFound, res.StatusCode, "Wrong status code")
	})

	t.Run("should return the same data when using timestamps with hours", func(t *testing.T) {

		n := runQuery(t, "en.wikipedia", "all-sites", "daily", 2021010100, 2021020100)

		assert.Len(t, n.Items, 31, "Unexpected response length")

		u := entities.UniqueDevices{
			Project:       "en.wikipedia",
			AccessSite:    "all-sites",
			Granularity:   "daily",
			Timestamp:     "20210102",
			Devices:       75002648,
			Offset:        14784457,
			Underestimate: 60218191,
		}

		assert.Equal(t, n.Items[0], u, "Wrong contents")

	})

	t.Run("should include offset and underestimate", func(t *testing.T) {

		n := runQuery(t, "en.wikipedia", "all-sites", "daily", 20210101, 20210201)

		assert.Equal(t, n.Items[0].Offset, 14784457, "Wrong contents")
		assert.Equal(t, n.Items[0].Underestimate, 60218191, "Wrong contents")

	})

	t.Run("should return numeric values as integers", func(t *testing.T) {

		n := runQuery(t, "en.wikipedia", "all-sites", "daily", 20210101, 20210201)

		assert.IsType(t, 0, n.Items[0].Devices, "Wrong type")
		assert.IsType(t, 0, n.Items[0].Offset, "Wrong type")
		assert.IsType(t, 0, n.Items[0].Underestimate, "Wrong type")

	})

}
