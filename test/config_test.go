package test

import (
	"fmt"
	"strings"
	"testing"

	"device-analytics/configuration"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaults(t *testing.T) {
	var config *configuration.Config
	var err error
	config, err = configuration.NewConfig([]byte{})
	require.NoError(t, err)
	require.NotNil(t, config)
	assert.Equal(t, "device-analytics", config.ServiceName)
	assert.Equal(t, "/metrics/unique-devices", config.BaseURI)
	assert.Equal(t, "localhost", config.Address)
	assert.Equal(t, 8080, config.Port)
	assert.Equal(t, "info", strings.ToLower(config.LogLevel))
	assert.Equal(t, 9042, config.Cassandra.Port)
	assert.Equal(t, "quorum", strings.ToLower(config.Cassandra.Consistency))
	assert.Len(t, config.Cassandra.Hosts, 1)
	assert.Equal(t, "localhost", config.Cassandra.Hosts[0])
}

func TestFullConfig(t *testing.T) {
	var err error
	var config *configuration.Config
	var conf string = `
service_name: test-service
base_uri: /v1
listen_address: 127.0.0.5
listen_port: 8081
log_level: debug
cassandra:
    port: 9043
    consistency: localQuorum
    hosts:
        - 127.0.0.6
        - 127.0.0.7
    local_dc: datacenter1
`
config, err = configuration.NewConfig([]byte(conf))
	require.NoError(t, err)
	require.NotNil(t, config)
	assert.Equal(t, "test-service", config.ServiceName)
	assert.Equal(t, "/v1", config.BaseURI)
	assert.Equal(t, "127.0.0.5", config.Address)
	assert.Equal(t, 8081, config.Port)
	assert.Equal(t, "debug", strings.ToLower(config.LogLevel))
	assert.Equal(t, 9043, config.Cassandra.Port)
	assert.Equal(t, "localquorum", strings.ToLower(config.Cassandra.Consistency))
	assert.Len(t, config.Cassandra.Hosts, 2)
	assert.Contains(t, config.Cassandra.Hosts, "127.0.0.6")
	assert.Contains(t, config.Cassandra.Hosts, "127.0.0.7")
	assert.Equal(t, "datacenter1", config.Cassandra.LocalDC)
}

func TestValidConsistencies(t *testing.T) {
	var conf = `
cassandra:
    consistency: %s
`
	var consistencyLevels = []string{
		"any",
		"one",
		"two",
		"three",
		"quorum",
		"all",
		"eachquorum",
		"localquorum",
		"localone",
		"QuOruM",
		"localONE",
	}
	for _, consistency := range consistencyLevels {
		t.Run(consistency, func(t *testing.T) {
			_, err := configuration.NewConfig([]byte(fmt.Sprintf(conf, consistency)))
			require.NoError(t, err)
		})
	}
}

func TestBogusConsistency(t *testing.T) {
	var conf string = `
cassandra:
    consistency: unreal
`
	_, err := configuration.NewConfig([]byte(conf))
	require.Error(t, err)
}

func TestValidLogLevels(t *testing.T) {
	for _, level := range []string{"debug", "info", "warning", "error", "fatal", "FaTaL", "INFO"} {
		t.Run(level, func(t *testing.T) {
			_, err := configuration.NewConfig([]byte(fmt.Sprintf("log_level: %s", level)))
			require.NoError(t, err)
		})
	}
}

func TestBogusLogLevel(t *testing.T) {
	_, err := configuration.NewConfig([]byte("log_level: unreal"))
	require.Error(t, err)
}
