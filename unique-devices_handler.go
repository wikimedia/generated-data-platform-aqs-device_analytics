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
	config  *Config
}

// API documentation
// @summary      Get unique devices per project
// @router       /unique-devices/{project}/{access-site}/{granularity}/{start}/{end}  [get]
// @description  Given a Wikimedia project and a date range, returns the number of unique devices that visited that wiki.
// @param        project      path  string  true  "Domain of a Wikimedia project"              example(en.wikipedia.org)
// @param        access-site  path  string  true  "Method of access"                           example(all-sites)  Enums(all-sites, desktop-site, mobile-site)
// @param        granularity  path  string  true  "Time unit for response data"                example(daily)  Enums(daily, monthly)
// @param        start        path  string  true  "First date to include, in YYYYMMDD format"  example(20220101)
// @param        end          path  string  true  "Last date to include, in YYYYMMDD format"   example(20220108)
// @produce      json
// @success      200  {object}  entities.UniqueDevicesResponse
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

	c, _ := context.WithTimeout(ctx, time.Duration(s.config.ContextTimeout)*time.Millisecond)
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
