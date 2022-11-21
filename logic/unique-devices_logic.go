package logic

import (
	"context"
	"net/http"
	"unique-devices/entities"

	"gitlab.wikimedia.org/frankie/aqsassist"
	"gerrit.wikimedia.org/r/mediawiki/services/servicelib-golang/logger"
	"github.com/gocql/gocql"
	"github.com/valyala/fasthttp"
	"schneider.vip/problem"
)

type UniqueDevicesLogic struct {
}

func (s *UniqueDevicesLogic) ProcessUniqueDevicesLogic(context context.Context, ctx *fasthttp.RequestCtx, project, accessSite, granularity, start, end string, session *gocql.Session, rLogger *logger.Logger) (*problem.Problem, entities.UniqueDevicesResponse) {
	var err error
	var problemData *problem.Problem
	var response = entities.UniqueDevicesResponse{Items: make([]entities.UniqueDevices, 0)}
	query := `SELECT devices, offset, underestimate, timestamp FROM "local_group_default_T_unique_devices".data WHERE "_domain" = 'analytics.wikimedia.org' AND project = ? AND "access-site" = ? AND granularity = ? AND timestamp >= ? AND timestamp <= ?`
	scanner := session.Query(query, project, accessSite, granularity, start, end).WithContext(context).Iter().Scanner()
	var devices, offset, underestimate int
	var timestamp string

	for scanner.Next() {
		if err = scanner.Scan(&devices, &offset, &underestimate, &timestamp); err != nil {
			rLogger.Log(logger.ERROR, "Query failed: %s", err)
			problemResp := aqsassist.CreateProblem(http.StatusInternalServerError, err.Error(), string(ctx.Request.URI().RequestURI())).JSON()
			ctx.SetStatusCode(http.StatusInternalServerError)
			ctx.SetBody(problemResp)
		}
		response.Items = append(response.Items, entities.UniqueDevices{
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
		problemResp := aqsassist.CreateProblem(http.StatusNotFound, str, string(ctx.Request.URI().RequestURI()))
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetBody(problemResp.JSON())
		return problemResp, entities.UniqueDevicesResponse{}
	}
	if err := scanner.Err(); err != nil {
		//s.logger.Request(r).Log(logger.ERROR, "Error querying database: %s", err)
		problemResp := aqsassist.CreateProblem(http.StatusInternalServerError, err.Error(), string(ctx.Request.URI().RequestURI()))
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBody(problemResp.JSON())
		return problemResp, entities.UniqueDevicesResponse{}
	}
	return problemData, response
}

