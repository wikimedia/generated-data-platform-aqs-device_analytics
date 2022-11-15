/*
 * Copyright 2021 Nikki Nikkhoui <nnikkhoui@wikimedia.org> and Wikimedia Foundation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"

	log "gerrit.wikimedia.org/r/mediawiki/services/servicelib-golang/logger"
	fasthttpprom "github.com/carousell/fasthttp-prometheus-middleware"
	"github.com/fasthttp/router"
	"github.com/gocql/gocql"
	"github.com/roger-russel/fasthttp-router-middleware/pkg/middleware"
	"github.com/valyala/fasthttp"
)

var (
	// These values are assigned at build using `-ldflags` (see: Makefile)
	buildDate = "unknown"
	buildHost = "unknown"
	version   = "unknown"
)

// Entrypoint for the service
func main() {
	var confFile = flag.String("config", "./config.yaml", "Path to the configuration file")

	var config *Config
	var err error
	var logger *log.Logger

	flag.Parse()

	if config, err = ReadConfig(*confFile); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	logger, err = log.NewLogger(os.Stdout, config.ServiceName, config.LogLevel)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to initialize the logger: %s", err)
		os.Exit(1)
	}

	logger.Info("Initializing service %s (Go version: %s, Build host: %s, Timestamp: %s", config.ServiceName, version, buildHost, buildDate)

	cluster := gocql.NewCluster(config.Address)
	cluster.Consistency = gocql.Quorum

	session, err := cluster.CreateSession()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	r := router.New()
	p := fasthttpprom.NewPrometheus("")
	p.MetricsPath = "/admin/metrics"
	p.Use(r)

	r.GET("/healthz", func(ctx *fasthttp.RequestCtx) {
		var response []byte
		ctx.SetStatusCode(fasthttp.StatusOK)
		if response, err = json.MarshalIndent(NewHealthz(version, buildDate, buildHost), "", "  "); err != nil {
			ctx.SetBody([]byte(`{}`))
			return
		}
		ctx.SetBody(response)
	})

	midAccessGroup := middleware.New([]middleware.Middleware{SetContentType, SecureHeadersMiddleware})

	// pass bound struct method to fasthttp
	uniqueDevicesHandler := &UniqueDevicesHandler{
		logger: logger, session: session}

	r.GET(path.Join(config.BaseURI, "/{project}/{access-site}/{granularity}/{start}/{end}"), midAccessGroup(uniqueDevicesHandler.HandleFastHTTP))

	fasthttp.ListenAndServe(fmt.Sprintf("%s:%d", config.Address, config.Port), r.Handler)
}
