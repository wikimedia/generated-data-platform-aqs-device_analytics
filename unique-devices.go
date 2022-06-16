package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gerrit.wikimedia.org/r/mediawiki/services/servicelib-golang/logger"
	"github.com/gocql/gocql"
	"github.com/julienschmidt/httprouter"
	"schneider.vip/problem"
)

//UniqueDevicesResponse represents a container for the unique devices resultset.
type UniqueDevicesResponse struct {
	Items []UniqueDevices `json:"items"`
}

//UniqueDevices represents one result from the unique devices resultset.
type UniqueDevices struct {
	Project     string `json:"project"`
	AccessSite  string `json:"access-site"`
	Granularity string `json:"granularity"`
	Timestamp   string `json:"timestamp"`
	Devices     int    `json:"devices"`
	Offset      int    `json:"offset"`
	Underestimate int `json:"underestimate"`
}

//UniqueDevicesHandler is the HTTP handler for unique devices API requests.
type UniqueDevicesHandler struct {
	logger  *logger.Logger
	session *gocql.Session
}

func (s *UniqueDevicesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	var params = httprouter.ParamsFromContext(r.Context())
	var response = UniqueDevicesResponse{Items: make([]UniqueDevices, 0)}

	project := TrimProjectDomain(params.ByName("project"))
	//strings.TrimPrefix(strings.TrimSuffix(strings.ToLower(params.ByName("project")), ".org"), "www.")
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

	if start, err = validateTimestamp(params.ByName("start")); err != nil {
		problem.New(
			problem.Type("about:blank"),
			problem.Title(http.StatusText(http.StatusBadRequest)),
			problem.Custom("method", http.MethodGet),
			problem.Status(http.StatusBadRequest),
			problem.Detail("Invalid timestamp"),
			problem.Custom("uri", r.RequestURI)).WriteTo(w)
		return
	}
	if end, err = validateTimestamp(params.ByName("end")); err != nil {
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
			Project:     project,
			AccessSite:  accessSite,
			Granularity: granularity,
			Timestamp:   timestamp,
			Devices:     devices,
			Offset:      offset,
			Underestimate: underestimate,
		})
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

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
	w.Write(data)
}

func validateTimestamp(param string) (string, error) {
	var err error
	var timestamp string

	// We accept timestamp parameters of two forms, YYYYMMDD and YYYYMMDDHH.
	// If timestamp parameter is 8 bytes length (8 ASCII runes),
	// then suffix the string with "00".

	if len(param) == 8 {
		timestamp = fmt.Sprintf("%s00", param)
	} else {
		timestamp = param
	}

	if _, err = time.Parse("2006010203", timestamp); err != nil {
		return "", err
	}

	return timestamp, nil
}

func TrimProjectDomain(param string) (string) {
	return strings.TrimPrefix(strings.TrimSuffix(strings.ToLower(param), ".org"), "www.")
}
