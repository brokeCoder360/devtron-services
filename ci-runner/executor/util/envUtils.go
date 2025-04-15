/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package util

import (
	"encoding/json"
	"fmt"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/ci-runner/helper"
	"github.com/devtron-labs/ci-runner/pubsub"
	"github.com/devtron-labs/ci-runner/util"
	"os"
	"strconv"
	"strings"
)

type ScriptEnvVariables struct {
	SystemEnv         map[string]string
	RuntimeEnv        map[string]string
	ExistingScriptEnv map[string]string
}

func (s *ScriptEnvVariables) ResetExistingScriptEnv() *ScriptEnvVariables {
	s.ExistingScriptEnv = make(map[string]string)
	return s
}

func getRuntimeEnvVariables(ciCdRequest *helper.CiCdTriggerEvent) map[string]string {
	if ciCdRequest.CommonWorkflowRequest.RuntimeEnvironmentVariables == nil {
		return make(map[string]string)
	}
	// setting runtime EnvironmentVariables
	return ciCdRequest.CommonWorkflowRequest.RuntimeEnvironmentVariables
}

func GetGlobalEnvVariables(ciCdRequest *helper.CiCdTriggerEvent) (*ScriptEnvVariables, error) {
	envs := make(map[string]string)
	envs["WORKING_DIRECTORY"] = util.WORKINGDIR
	cfg := &pubsub.PubSubConfig{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}
	if helper.IsCIOrJobTypeEvent(ciCdRequest.Type) {
		image, err := helper.BuildDockerImagePath(ciCdRequest.CommonWorkflowRequest)
		if err != nil {
			return nil, err
		}

		envs["DOCKER_IMAGE_TAG"] = ciCdRequest.CommonWorkflowRequest.DockerImageTag
		envs["DOCKER_REPOSITORY"] = ciCdRequest.CommonWorkflowRequest.DockerRepository
		envs["DOCKER_REGISTRY_URL"] = ciCdRequest.CommonWorkflowRequest.DockerRegistryURL
		envs["TRIGGER_BY_AUTHOR"] = ciCdRequest.CommonWorkflowRequest.TriggerByAuthor
		envs["DOCKER_IMAGE"] = image
		envs["DOCKER_IMAGE_TAG"] = ciCdRequest.CommonWorkflowRequest.DockerImageTag

		if ciCdRequest.Type == util.JOBEVENT {
			envs["JOB_NAME"] = ciCdRequest.CommonWorkflowRequest.AppName
		} else {
			envs["APP_NAME"] = ciCdRequest.CommonWorkflowRequest.AppName
		}
		//adding GIT_MATERIAL_REQUEST in env for semgrep plugin
		CiMaterialRequestArr := ""
		if ciCdRequest.CommonWorkflowRequest.CiProjectDetails != nil {
			for _, ciProjectDetail := range ciCdRequest.CommonWorkflowRequest.CiProjectDetails {
				GitRepoSplit := strings.Split(ciProjectDetail.GitRepository, "/")
				GitRepoName := ""
				if len(GitRepoSplit) > 0 {
					GitRepoName = strings.Split(GitRepoSplit[len(GitRepoSplit)-1], ".")[0]
				}
				CiMaterialRequestArr = CiMaterialRequestArr +
					fmt.Sprintf("%s,%s,%s,%s|", GitRepoName, ciProjectDetail.CheckoutPath, ciProjectDetail.SourceValue, ciProjectDetail.CommitHash)
			}
		}
		envs["GIT_MATERIAL_REQUEST"] = CiMaterialRequestArr // GIT_MATERIAL_REQUEST will be of form "<repoName>/<checkoutPath>/<BranchName>/<CommitHash>"
		fmt.Println(envs["GIT_MATERIAL_REQUEST"])

		// adding envs for polling-plugin
		envs["DOCKER_REGISTRY_TYPE"] = ciCdRequest.CommonWorkflowRequest.DockerRegistryType
		envs["DOCKER_USERNAME"] = ciCdRequest.CommonWorkflowRequest.DockerUsername
		envs["DOCKER_PASSWORD"] = ciCdRequest.CommonWorkflowRequest.DockerPassword
		envs["ACCESS_KEY"] = ciCdRequest.CommonWorkflowRequest.AccessKey
		envs["SECRET_KEY"] = ciCdRequest.CommonWorkflowRequest.SecretKey
		envs["AWS_REGION"] = ciCdRequest.CommonWorkflowRequest.AwsRegion
		envs["LAST_FETCHED_TIME"] = ciCdRequest.CommonWorkflowRequest.CiArtifactLastFetch.String()

		//adding some envs for Image scanning plugin
		envs["PIPELINE_ID"] = strconv.Itoa(ciCdRequest.CommonWorkflowRequest.PipelineId)
		envs["TRIGGERED_BY"] = strconv.Itoa(ciCdRequest.CommonWorkflowRequest.TriggeredBy)
		envs["DOCKER_REGISTRY_ID"] = ciCdRequest.CommonWorkflowRequest.DockerRegistryId
		envs["IMAGE_SCANNER_ENDPOINT"] = cfg.ImageScannerEndpoint
		envs["IMAGE_SCAN_MAX_RETRIES"] = strconv.Itoa(ciCdRequest.CommonWorkflowRequest.ImageScanMaxRetries)
		envs["IMAGE_SCAN_RETRY_DELAY"] = strconv.Itoa(ciCdRequest.CommonWorkflowRequest.ImageScanRetryDelay)

		// setting system EnvironmentVariables
		for k, v := range ciCdRequest.CommonWorkflowRequest.SystemEnvironmentVariables {
			envs[k] = v
		}
		// for skopeo plugin, list of destination images againt registry name eg: <registry_name>: [<i1>,<i2>]
		RegistryDestinationImage, _ := json.Marshal(ciCdRequest.CommonWorkflowRequest.RegistryDestinationImageMap)
		RegistryCredentials, _ := json.Marshal(ciCdRequest.CommonWorkflowRequest.RegistryCredentialMap)
		envs["REGISTRY_DESTINATION_IMAGE_MAP"] = string(RegistryDestinationImage)
		envs["REGISTRY_CREDENTIALS"] = string(RegistryCredentials)
		envs["AWS_INSPECTOR_CONFIG"] = ciCdRequest.CommonWorkflowRequest.AwsInspectorConfig
	} else {
		envs["DOCKER_IMAGE"] = ciCdRequest.CommonWorkflowRequest.CiArtifactDTO.Image
		envs["DOCKER_IMAGE_TAG"] = ciCdRequest.CommonWorkflowRequest.DockerImageTag
		envs["DEPLOYMENT_RELEASE_ID"] = strconv.Itoa(ciCdRequest.CommonWorkflowRequest.DeploymentReleaseCounter)
		envs["DEPLOYMENT_UNIQUE_ID"] = strconv.Itoa(ciCdRequest.CommonWorkflowRequest.WorkflowRunnerId)
		envs["CD_TRIGGERED_BY"] = ciCdRequest.CommonWorkflowRequest.DeploymentTriggeredBy
		envs["CD_TRIGGER_TIME"] = ciCdRequest.CommonWorkflowRequest.DeploymentTriggerTime.String()

		// to support legacy yaml based script trigger
		envs["DEVTRON_CD_TRIGGERED_BY"] = ciCdRequest.CommonWorkflowRequest.DeploymentTriggeredBy
		envs["DEVTRON_CD_TRIGGER_TIME"] = ciCdRequest.CommonWorkflowRequest.DeploymentTriggerTime.String()

		//adding some envs for Image scanning plugin
		envs["TRIGGERED_BY"] = strconv.Itoa(ciCdRequest.CommonWorkflowRequest.TriggeredBy)
		envs["DOCKER_REGISTRY_ID"] = ciCdRequest.CommonWorkflowRequest.DockerRegistryId
		envs["IMAGE_SCANNER_ENDPOINT"] = cfg.ImageScannerEndpoint
		envs["IMAGE_SCAN_MAX_RETRIES"] = strconv.Itoa(ciCdRequest.CommonWorkflowRequest.ImageScanMaxRetries)
		envs["IMAGE_SCAN_RETRY_DELAY"] = strconv.Itoa(ciCdRequest.CommonWorkflowRequest.ImageScanRetryDelay)

		// setting system EnvironmentVariables
		for k, v := range ciCdRequest.CommonWorkflowRequest.SystemEnvironmentVariables {
			envs[k] = v
		}
		// for skopeo plugin, list of destination images against registry name eg: <registry_name>: [<i1>,<i2>]
		RegistryDestinationImage, _ := json.Marshal(ciCdRequest.CommonWorkflowRequest.RegistryDestinationImageMap)
		RegistryCredentials, _ := json.Marshal(ciCdRequest.CommonWorkflowRequest.RegistryCredentialMap)
		envs["REGISTRY_DESTINATION_IMAGE_MAP"] = string(RegistryDestinationImage)
		envs["REGISTRY_CREDENTIALS"] = string(RegistryCredentials)
	}
	scriptEnvVariables := &ScriptEnvVariables{
		SystemEnv:  envs,
		RuntimeEnv: getRuntimeEnvVariables(ciCdRequest),
	}
	return scriptEnvVariables, nil
}

func GetSystemEnvVariables() map[string]string {
	envs := make(map[string]string)
	//get all environment variables
	envVars := os.Environ()
	for _, envVar := range envVars {
		subs := strings.SplitN(envVar, "=", 2)
		if len(subs) != 2 {
			// skip invalid env variables for panic handling
			continue
		}
		// TODO: We're currently using CI_CD_EVENT in our pre-defined Plugins.
		//  Remove this dependency and then remove add this check
		//if subs[0] == util.CiCdEventEnvKey {
		//	// skip CI_CD_EVENT env variable as it is internal to the system
		//	continue
		//}
		envs[subs[0]] = subs[1]
	}
	return envs
}
