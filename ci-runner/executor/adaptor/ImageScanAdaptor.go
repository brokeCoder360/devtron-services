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

package adaptor

import (
	bean2 "github.com/devtron-labs/ci-runner/executor/stage/bean"
	util2 "github.com/devtron-labs/ci-runner/executor/util"
	"github.com/devtron-labs/ci-runner/helper"
	"github.com/devtron-labs/common-lib/constants"
	"github.com/devtron-labs/common-lib/imageScan/bean"
)

func GetImageScanEvent(dest, digest string, commonWorkflowRequest *helper.CommonWorkflowRequest) *helper.ScanEvent {
	if commonWorkflowRequest == nil {
		return &helper.ScanEvent{}
	}
	return &helper.ScanEvent{
		ImageScanEvent: bean.ImageScanEvent{
			Image:            dest,
			ImageDigest:      digest,
			PipelineId:       commonWorkflowRequest.PipelineId,
			UserId:           commonWorkflowRequest.TriggeredBy,
			DockerRegistryId: commonWorkflowRequest.DockerRegistryId,
			DockerConnection: commonWorkflowRequest.DockerConnection,
			DockerCert:       commonWorkflowRequest.DockerCert,
			SourceType:       constants.SourceTypeImage,
			SourceSubType:    constants.SourceSubTypeCi,
		},
		ImageScanMaxRetries: commonWorkflowRequest.ImageScanMaxRetries,
		ImageScanRetryDelay: commonWorkflowRequest.ImageScanRetryDelay,
	}
}
func GetImageScannerExecutorBean(ciCdRequest *helper.CiCdTriggerEvent, scriptEnvs *util2.ScriptEnvVariables, refStageMap map[int][]*helper.StepObject, metrics *helper.CIMetrics, artifactUploaded bool, dest string, digest string) *bean2.ImageScanningExecutorBean {
	return &bean2.ImageScanningExecutorBean{
		CiCdRequest:      ciCdRequest,
		ScriptEnvs:       scriptEnvs,
		RefStageMap:      refStageMap,
		Metrics:          metrics,
		ArtifactUploaded: artifactUploaded,
		Dest:             dest,
		Digest:           digest,
	}
}
