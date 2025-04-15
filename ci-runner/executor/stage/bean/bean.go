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

package bean

import (
	util2 "github.com/devtron-labs/ci-runner/executor/util"
	"github.com/devtron-labs/ci-runner/helper"
)

const (
	ExternalCiArtifact = "externalCiArtifact"
	ImageDigest        = "imageDigest"
	UseAppDockerConfig = "useAppDockerConfig"
	CiProjectDetails   = "ciProjectDetails"
)
const (
	DigestGlobalEnvKey     = "DIGEST"
	ScanToolIdGlobalEnvKey = "SCAN_TOOL_ID"
)

type ImageScanningExecutorBean struct {
	CiCdRequest      *helper.CiCdTriggerEvent
	ScriptEnvs       *util2.ScriptEnvVariables
	RefStageMap      map[int][]*helper.StepObject
	Metrics          *helper.CIMetrics
	ArtifactUploaded bool
	Dest             string
	Digest           string
}
