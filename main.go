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
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"
	"runtime"

	log "gerrit.wikimedia.org/r/mediawiki/services/servicelib-golang/logger"
	"github.com/eevans/servicelib-golang/middleware"
	"github.com/gocql/gocql"
	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// These values are assigned at build using `-ldflags` (see: Makefile)
	buildDate = "unknown"
	buildHost = "unknown"
	version   = "unknown"
)

var (
	reqCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Count of HTTP requests processed, partitioned by status code and HTTP method.",
		},
		[]string{"code", "method"},
	)

	durationHisto = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "A histogram of latencies for requests, partitioned by status code and HTTP method.",
			Buckets: []float64{.001, .0025, .0050, .01, .025, .050, .10, .25, .50, 1},
		},
		[]string{"code", "method"},
	)
	promBuildInfoGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        "service_scaffold_golang_build_info",
			Help:        "Build information",
			ConstLabels: map[string]string{"version": version, "build_date": buildDate, "build_host": buildHost, "go_version": runtime.Version()},
		})
)

func init() {
	prometheus.MustRegister(reqCounter, durationHisto, promBuildInfoGauge)
	promBuildInfoGauge.Set(1)
}

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

	// HTTP handler wrapped in Prometheus middleware
	uniqueDevicesHandler := middleware.PrometheusInstrumentationMiddleware(reqCounter, durationHisto)(&UniqueDevicesHandler{logger: logger, session: session})

	router := httprouter.New()
	router.Handler("GET", "/healthz", &HealthzHandler{NewHealthz(version, buildDate, buildHost)})
	router.Handler("GET", "/admin/metrics", promhttp.Handler())
	router.Handler("GET", path.Join(config.BaseURI, "/:project/:access-site/:granularity/:start/:end"), uniqueDevicesHandler)

	http.ListenAndServe(fmt.Sprintf("%s:%d", config.Address, config.Port), router)
}
