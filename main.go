/*
 * Copyright 2022 Wikimedia Foundation
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
	"strings"

	log "gerrit.wikimedia.org/r/mediawiki/services/servicelib-golang/logger"
	fasthttpprom "github.com/carousell/fasthttp-prometheus-middleware"
	"github.com/fasthttp/router"
	"github.com/roger-russel/fasthttp-router-middleware/pkg/middleware"
	"github.com/valyala/fasthttp"
)

var (
	// These values are assigned at build using `-ldflags` (see: Makefile)
	buildDate = "unknown"
	buildHost = "unknown"
	version   = "unknown"
)

// API documentation
// @title                 Wikimedia Unique Devices API
// @version               DRAFT
// @description.markdown  api.md
// @contact.name
// @contact.url
// @contact.email
// @license.name          Apache 2.0
// @license.url           http://www.apache.org/licenses/LICENSE-2.0.html
// @termsOfService        https://wikimediafoundation.org/wiki/Terms_of_Use
// @host                  localhost:8080
// @basePath              /metrics/
// @schemes               http

// Entrypoint for the service
func main() {
	var confFile = flag.String("config", "./config.yaml", "Path to the configuration file")

	var config *Config
	var err error
	var logger *log.Logger

	notFoundHandler := &NotFoundHandler{}

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

	logger.Info("Connecting to Cassandra database(s): %s (port %d)", strings.Join(config.Cassandra.Hosts, ","), config.Cassandra.Port)
	logger.Debug("Cassandra: configured for consistency level '%s'", strings.ToLower(config.Cassandra.Consistency))
	logger.Debug("Cassandra: configured for local datacenter '%s'", config.Cassandra.LocalDC)

	session, err := newCassandraSession(config)
	if err != nil {
		logger.Error("Failed to create Cassandra session: %s", err)
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// pass bound struct method to fasthttp
	uniqueDevicesHandler := &UniqueDevicesHandler{
		logger: logger, session: session, config: config}

	r := router.New()
	r.NotFound = notFoundHandler.HandleFastHTTP
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

	r.GET(path.Join(config.BaseURI, "/{project}/{access-site}/{granularity}/{start}/{end}"), midAccessGroup(uniqueDevicesHandler.HandleFastHTTP))

	err = fasthttp.ListenAndServe(fmt.Sprintf("%s:%d", config.Address, config.Port), r.Handler)
	logger.Info(err.Error())
}
