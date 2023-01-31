/*
 * Copyright 2022 Eric Evans <eevans@wikimedia.org> and Wikimedia Foundation
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
	"fmt"
	"strings"

	"github.com/gocql/gocql"
)

// Return a new Cassandra session corresponding to the provided config.
func newCassandraSession(config *Config) (*gocql.Session, error) {
	var cluster *gocql.ClusterConfig = gocql.NewCluster(config.Cassandra.Hosts...)

	cluster.Consistency, _ = goCQLConsistency(config.Cassandra.Consistency)
	cluster.Port = config.Cassandra.Port

	// Host selection
	if config.Cassandra.LocalDC != "" {
		cluster.PoolConfig.HostSelectionPolicy = gocql.DCAwareRoundRobinPolicy(config.Cassandra.LocalDC)
	} else {
		cluster.PoolConfig.HostSelectionPolicy = gocql.RoundRobinHostPolicy()
	}

	return cluster.CreateSession()
}

// Given a string, return the corresponding GoCQL consistency level type.
func goCQLConsistency(c string) (gocql.Consistency, error) {
	switch strings.ToLower(c) {
	case "any":
		return gocql.Any, nil
	case "one":
		return gocql.One, nil
	case "two":
		return gocql.Two, nil
	case "three":
		return gocql.Three, nil
	case "quorum":
		return gocql.Quorum, nil
	case "all":
		return gocql.All, nil
	case "localquorum":
		return gocql.LocalQuorum, nil
	case "eachquorum":
		return gocql.EachQuorum, nil
	case "localone":
		return gocql.LocalOne, nil
	default:
		return 0, fmt.Errorf("Unrecognized Cassandra consistency level")
	}
}
