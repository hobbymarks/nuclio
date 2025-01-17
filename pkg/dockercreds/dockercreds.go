/*
Copyright 2023 The Nuclio Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package dockercreds

import (
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/nuclio/nuclio/pkg/common"
	"github.com/nuclio/nuclio/pkg/dockerclient"

	"github.com/nuclio/errors"
	"github.com/nuclio/logger"
)

type Credentials struct {
	Username        string
	Password        string
	URL             string
	RefreshInterval *time.Duration
}

// DockerCreds initializes docker client credentials
type DockerCreds struct {
	logger          logger.Logger
	dockerClient    dockerclient.Client
	refreshInterval *time.Duration
	dockerCreds     []*dockerCred
}

func NewDockerCreds(parentLogger logger.Logger,
	dockerClient dockerclient.Client,
	refreshInterval *time.Duration) (*DockerCreds, error) {

	return &DockerCreds{
		logger:          parentLogger.GetChild("dockercreds"),
		dockerClient:    dockerClient,
		refreshInterval: refreshInterval,
	}, nil
}

func (dc *DockerCreds) LoadFromDir(keyDir string) error {
	dockerKeyFileInfos, err := os.ReadDir(keyDir)
	if err != nil {
		return errors.Wrap(err, "Failed to read docker key directory")
	}

	for _, dockerKeyFileInfo := range dockerKeyFileInfos {

		// create the full path of the docker credentials
		dockerKeyFilePath := path.Join(keyDir, dockerKeyFileInfo.Name())

		// evaluate just in case it's a symlink
		dockerKeyFilePath, err = filepath.EvalSymlinks(dockerKeyFilePath)
		if err != nil {
			dc.logger.WarnWith("Failed to evaluate symlink",
				"err", errors.Cause(err),
				"path", dockerKeyFilePath)

			continue
		}

		if common.IsDir(dockerKeyFilePath) {
			continue
		}

		dockerCred, err := newDockerCred(dc, dockerKeyFilePath, dc.refreshInterval)
		if err != nil {
			dc.logger.WarnWith("Failed to create docker cred", "err", err)
			continue
		}

		dc.dockerCreds = append(dc.dockerCreds, dockerCred)
	}

	return nil
}

func (dc *DockerCreds) GetCredentials() []Credentials {
	var credentials []Credentials

	for _, dockerCred := range dc.dockerCreds {
		credentials = append(credentials, dockerCred.credentials)
	}

	return credentials
}
