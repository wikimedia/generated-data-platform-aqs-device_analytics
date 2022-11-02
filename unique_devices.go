package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"gerrit.wikimedia.org/r/mediawiki/services/servicelib-golang/logger"
	"github.com/gocql/gocql"
	"github.com/julienschmidt/httprouter"
	"gitlab.wikimedia.org/frankie/aqsassist"
	"schneider.vip/problem"
)

// UniqueDevicesResponse represents a container for the unique devices resultset.
type UniqueDevicesResponse struct {
	Items []UniqueDevices `json:"items"`
}

// UniqueDevices represents one result from the unique devices resultset.
type UniqueDevices struct {
	Project       string `json:"project"`
	AccessSite    string `json:"access-site"`
	Granularity   string `json:"granularity"`
	Timestamp     string `json:"timestamp"`
	Devices       int    `json:"devices"`
	Offset        int    `json:"offset"`
	Underestimate int    `json:"underestimate"`
}

// UniqueDevicesHandler is the HTTP handler for unique devices API requests.
type UniqueDevicesHandler struct {
	logger  *logger.Logger
	session *gocql.Session
}

func (s *UniqueDevicesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	var err error
	var params = httprouter.ParamsFromContext(r.Context())
	var response = UniqueDevicesResponse{Items: make([]UniqueDevices, 0)}

	project := aqsassist.TrimProjectDomain(params.ByName("project"))
	accessSite := strings.ToLower(params.ByName("access-site"))
	granularity := strings.ToLower(params.ByName("granularity"))
	var start, end string

	if granularity != "daily" && granularity != "monthly" && granularity != "hourly" {
		problem.New(
			problem.Type("about:blank"),
			problem.Title(http.StatusText(http.StatusBadRequest)),
			problem.Custom("method", http.MethodGet),
			problem.Status(http.StatusBadRequest),
			problem.Detail("Invalid granularity"),
			problem.Custom("uri", r.RequestURI)).WriteTo(w)
		return
	}

	if start, err = aqsassist.ValidateTimestamp(params.ByName("start")); err != nil {
		problem.New(
			problem.Type("about:blank"),
			problem.Title(http.StatusText(http.StatusBadRequest)),
			problem.Custom("method", http.MethodGet),
			problem.Status(http.StatusBadRequest),
			problem.Detail("Invalid timestamp"),
			problem.Custom("uri", r.RequestURI)).WriteTo(w)
		return
	}

	if end, err = aqsassist.ValidateTimestamp(params.ByName("end")); err != nil {
		problem.New(
			problem.Type("about:blank"),
			problem.Title(http.StatusText(http.StatusBadRequest)),
			problem.Custom("method", http.MethodGet),
			problem.Status(http.StatusBadRequest),
			problem.Detail("Invalid timestamp"),
			problem.Custom("uri", r.RequestURI)).WriteTo(w)
		return
	}

	if err = aqsassist.StartBeforeEnd(start, end); err != nil {
		problem.New(
			problem.Type("about:blank"),
			problem.Title(http.StatusText(http.StatusBadRequest)),
			problem.Custom("method", http.MethodGet),
			problem.Status(http.StatusBadRequest),
			problem.Detail("Invalid timestamp"),
			problem.Custom("uri", r.RequestURI)).WriteTo(w)
		return
	}

	ctx := context.Background()

	query := `SELECT devices, offset, underestimate, timestamp FROM "local_group_default_T_unique_devices".data WHERE "_domain" = 'analytics.wikimedia.org' AND project = ? AND "access-site" = ? AND granularity = ? AND timestamp >= ? AND timestamp <= ?`
	scanner := s.session.Query(query, project, accessSite, granularity, start, end).WithContext(ctx).Iter().Scanner()
	var devices, offset, underestimate int
	var timestamp string

	for scanner.Next() {
		if err = scanner.Scan(&devices, &offset, &underestimate, &timestamp); err != nil {
			s.logger.Request(r).Log(logger.ERROR, "Query failed: %s", err)
			problem.New(
				problem.Type("about:blank"),
				problem.Title(http.StatusText(http.StatusInternalServerError)),
				problem.Custom("method", http.MethodGet),
				problem.Status(http.StatusInternalServerError),
				problem.Detail(err.Error()),
				problem.Custom("uri", r.RequestURI)).WriteTo(w)
		}
		response.Items = append(response.Items, UniqueDevices{
			Project:       project,
			AccessSite:    accessSite,
			Granularity:   granularity,
			Timestamp:     timestamp,
			Devices:       devices,
			Offset:        offset,
			Underestimate: underestimate,
		})
	}

	str := "The date(s) you used are valid, but we either do not have data for those date(s), or the project you asked for is not loaded yet.  Please check documentation for more information."
	if len(response.Items) == 0 {
		aqsassist.HandleEmptyResponse(w, r, str)
		return
	}

	if err := scanner.Err(); err != nil {
		s.logger.Request(r).Log(logger.ERROR, "Error querying database: %s", err)
		problem.New(
			problem.Type("about:blank"),
			problem.Title(http.StatusText(http.StatusInternalServerError)),
			problem.Custom("method", http.MethodGet),
			problem.Status(http.StatusInternalServerError),
			problem.Detail(err.Error()),
			problem.Custom("uri", r.RequestURI)).WriteTo(w)
		return
	}

	var data []byte
	if data, err = json.MarshalIndent(response, "", " "); err != nil {
		s.logger.Request(r).Log(logger.ERROR, "Unable to marshal response object: %s", err)
		problem.New(
			problem.Type("about:blank"),
			problem.Title(http.StatusText(http.StatusInternalServerError)),
			problem.Custom("method", http.MethodGet),
			problem.Status(http.StatusInternalServerError),
			problem.Detail(err.Error()),
			problem.Custom("uri", r.RequestURI)).WriteTo(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
