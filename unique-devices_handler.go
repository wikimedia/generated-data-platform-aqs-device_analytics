package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"
	"unique-devices/logic"

	"gerrit.wikimedia.org/r/mediawiki/services/servicelib-golang/logger"
	"github.com/gocql/gocql"
	"github.com/valyala/fasthttp"
	"gitlab.wikimedia.org/frankie/aqsassist"
)

// UniqueDevicesHandler is the HTTP handler for unique devices API requests.
type UniqueDevicesHandler struct {
	logger  *logger.Logger
	session *gocql.Session
	logic   *logic.UniqueDevicesLogic
}

func (s *UniqueDevicesHandler) HandleFastHTTP(ctx *fasthttp.RequestCtx) {
	var err error

	project := aqsassist.TrimProjectDomain(ctx.UserValue("project").(string))
	accessSite := strings.ToLower(ctx.UserValue("access-site").(string))
	granularity := strings.ToLower(ctx.UserValue("granularity").(string))
	var start, end string

	if granularity != "daily" && granularity != "monthly" && granularity != "hourly" {
		problemResp := aqsassist.CreateProblem(http.StatusBadRequest, "Invalid granularity", string(ctx.Request.URI().RequestURI())).JSON()
		ctx.SetStatusCode(http.StatusBadRequest)
		ctx.SetBody(problemResp)
		return
	}

	if start, err = aqsassist.ValidateTimestamp(ctx.UserValue("start").(string)); err != nil {
		problemResp := aqsassist.CreateProblem(http.StatusBadRequest, "start timestamp is invalid, must be a valid date in YYYYMMDD format", string(ctx.Request.URI().RequestURI())).JSON()
		ctx.SetStatusCode(http.StatusBadRequest)
		ctx.SetBody(problemResp)
		return
	}
	if end, err = aqsassist.ValidateTimestamp(ctx.UserValue("end").(string)); err != nil {
		problemResp := aqsassist.CreateProblem(http.StatusBadRequest, "end timestamp is invalid, must be a valid date in YYYYMMDD format", string(ctx.Request.URI().RequestURI())).JSON()
		ctx.SetStatusCode(http.StatusBadRequest)
		ctx.SetBody(problemResp)
		return
	}

	if err = aqsassist.StartBeforeEnd(start, end); err != nil {
		problemResp := aqsassist.CreateProblem(http.StatusBadRequest, err.Error(), string(ctx.Request.URI().RequestURI())).JSON()
		ctx.SetStatusCode(http.StatusBadRequest)
		ctx.SetBody(problemResp)
		return
	}

	c, _ := context.WithTimeout(ctx, 40*time.Millisecond)
	pbm, response := s.logic.ProcessUniqueDevicesLogic(c, ctx, project, accessSite, granularity, start, end, s.session, s.logger)
	if pbm != nil {
		problemResp, _ := json.Marshal(pbm)
		ctx.SetBody(problemResp)
		return
	}

	var data []byte
	if data, err = json.MarshalIndent(response, "", " "); err != nil {
		s.logger.Log(logger.ERROR, "Unable to marshal response object: %s", err)
		problemResp := aqsassist.CreateProblem(http.StatusInternalServerError, err.Error(), string(ctx.Request.URI().RequestURI())).JSON()
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBody(problemResp)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBody([]byte(data))
}
