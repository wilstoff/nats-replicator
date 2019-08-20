/*
 * Copyright 2019 The NATS Authors
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package core

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/nats-io/nats-replicator/server/conf"
	"github.com/stretchr/testify/require"
)

func TestStartWithConfigFileFlag(t *testing.T) {
	tbs, err := StartTestEnvironmentInfrastructure(false)
	require.NoError(t, err)
	defer tbs.Close()

	file, err := ioutil.TempFile(os.TempDir(), "config")
	require.NoError(t, err)

	configString := `
	{
		connectors: [],
		nats: [
			{
				name: "one"
				servers: ["%s"]
				norandom: true
				no_echo: true
			}
		]
		monitoring: {
			HTTPPort: -1,
			ReadTimeout: 2000,
		}
	}
	`
	configString = fmt.Sprintf(configString, tbs.natsURL)

	fullPath, err := conf.ValidateFilePath(file.Name())
	require.NoError(t, err)

	err = ioutil.WriteFile(fullPath, []byte(configString), 0644)
	require.NoError(t, err)

	flags := Flags{
		ConfigFile:      fullPath,
		DebugAndVerbose: true,
	}

	server := NewNATSReplicator()
	server.InitializeFromFlags(flags)
	err = server.Start()
	require.NoError(t, err)
	defer server.Stop()

	require.Equal(t, 1, len(server.config.NATS))
	require.Equal(t, 1, len(server.config.NATS[0].Servers))
	require.Equal(t, server.config.NATS[0].Servers[0], tbs.natsURL)
	require.Equal(t, server.config.NATS[0].NoRandom, true)
	require.Equal(t, server.config.NATS[0].NoEcho, true)
	require.Equal(t, server.config.Monitoring.ReadTimeout, 2000)
	require.Equal(t, server.config.Logging.Trace, true)
	require.Equal(t, server.config.Logging.Debug, true)

	httpClient := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := httpClient.Get(fmt.Sprintf("%s/healthz", server.monitoringURL))
	require.NoError(t, err)
	require.True(t, resp.StatusCode == http.StatusOK)
}

func TestStartWithConfigFileEnv(t *testing.T) {
	tbs, err := StartTestEnvironmentInfrastructure(false)
	require.NoError(t, err)
	defer tbs.Close()

	file, err := ioutil.TempFile(os.TempDir(), "config")
	require.NoError(t, err)

	configString := `
	{
		connectors: [],
		nats: [
			{
				name: "one"
				servers: ["%s"]
			}
		]
		monitoring: {
			HTTPPort: -1,
			ReadTimeout: 2000,
		}
	}
	`
	configString = fmt.Sprintf(configString, tbs.natsURL)

	fullPath, err := conf.ValidateFilePath(file.Name())
	require.NoError(t, err)

	err = ioutil.WriteFile(fullPath, []byte(configString), 0644)
	require.NoError(t, err)

	flags := Flags{
		ConfigFile:      "",
		DebugAndVerbose: true,
	}

	os.Setenv("NATS_REPLICATOR_CONFIG", fullPath)
	server := NewNATSReplicator()
	server.InitializeFromFlags(flags)
	err = server.Start()
	require.NoError(t, err)
	defer server.Stop()
	os.Setenv("NATS_REPLICATOR_CONFIG", "")

	require.Equal(t, 1, len(server.config.NATS))
	require.Equal(t, 1, len(server.config.NATS[0].Servers))
	require.Equal(t, server.config.NATS[0].Servers[0], tbs.natsURL)
	require.Equal(t, server.config.Monitoring.ReadTimeout, 2000)
	require.Equal(t, server.config.Logging.Trace, true)
	require.Equal(t, server.config.Logging.Debug, true)

	httpClient := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := httpClient.Get(fmt.Sprintf("%s/healthz", server.monitoringURL))
	require.NoError(t, err)
	require.True(t, resp.StatusCode == http.StatusOK)
}

func TestFailWithoutConfigFile(t *testing.T) {
	tbs, err := StartTestEnvironmentInfrastructure(false)
	require.NoError(t, err)
	defer tbs.Close()

	flags := Flags{
		ConfigFile:      "",
		DebugAndVerbose: true,
	}

	os.Setenv("NATS_REPLICATOR_CONFIG", "")
	server := NewNATSReplicator()
	err = server.InitializeFromFlags(flags)
	require.Error(t, err)
}
