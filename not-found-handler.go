package main

import (
	"net/http"

	"github.com/valyala/fasthttp"
	"gitlab.wikimedia.org/frankie/aqsassist"
)

// NotFoundHandler is the HTTP handler when no match routes are found.
type NotFoundHandler struct {
}

func (s *NotFoundHandler) HandleFastHTTP(ctx *fasthttp.RequestCtx) {
	problemResp := aqsassist.CreateProblem(http.StatusNotFound, "Invalid route", string(ctx.Request.URI().RequestURI())).JSON()
	ctx.SetStatusCode(http.StatusBadRequest)
	ctx.SetBody(problemResp)
}
