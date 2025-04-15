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

package stage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/devtron-labs/ci-runner/executor"
	adaptor2 "github.com/devtron-labs/ci-runner/executor/adaptor"
	cicxt "github.com/devtron-labs/ci-runner/executor/context"
	helper2 "github.com/devtron-labs/ci-runner/executor/helper"
	bean2 "github.com/devtron-labs/ci-runner/executor/stage/bean"
	util2 "github.com/devtron-labs/ci-runner/executor/util"
	"github.com/devtron-labs/ci-runner/helper"
	"github.com/devtron-labs/ci-runner/helper/adaptor"
	"github.com/devtron-labs/ci-runner/util"
	"github.com/devtron-labs/common-lib/utils"
	"github.com/devtron-labs/common-lib/utils/bean"
	"github.com/devtron-labs/common-lib/utils/workFlow"
	commonBean "github.com/devtron-labs/common-lib/workflow"
	"golang.org/x/sync/errgroup"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

/*
 *  Copyright 2020 Devtron Labs
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
 *
 */

type CiStage struct {
	gitManager           helper.GitManager
	dockerHelper         helper.DockerHelper
	stageExecutorManager executor.StageExecutor
}

func NewCiStage(gitManager helper.GitManager, dockerHelper helper.DockerHelper, stageExecutor executor.StageExecutor) *CiStage {
	return &CiStage{
		gitManager:           gitManager,
		dockerHelper:         dockerHelper,
		stageExecutorManager: stageExecutor,
	}
}

func deferCIEvent(ciRequest *helper.CommonWorkflowRequest, artifactUploaded bool, err error) (exitCode int) {
	log.Println(util.DEVTRON, "defer CI stage data.", "err: ", err, "artifactUploaded: ", artifactUploaded)
	if err != nil {
		var stageError *helper.CiStageError
		if errors.As(err, &stageError) {
			exitCode = workFlow.CiStageFailErrorCode
			// update artifact uploaded status
			if !stageError.IsArtifactUploaded() {
				stageError = stageError.WithArtifactUploaded(artifactUploaded)
			}
		} else {
			exitCode = workFlow.DefaultErrorCode
			stageError = helper.NewCiStageError(err).
				WithArtifactUploaded(artifactUploaded).
				WithFailureMessage(workFlow.CiFailed.String())
		}
		// send ci failure event, for ci failure notification
		sendCIFailureEvent(ciRequest, stageError)
		// populate stage error
		util.PopulateStageError(stageError.ErrorMessage())
	}
	return exitCode
}

func (impl *CiStage) HandleCIEvent(ciCdRequest *helper.CiCdTriggerEvent, exitCode *int) {
	var artifactUploaded bool
	var err error
	ciRequest := ciCdRequest.CommonWorkflowRequest
	ciContext := cicxt.BuildCiContext(context.Background(), ciRequest.EnableSecretMasking)
	defer func() {
		*exitCode = deferCIEvent(ciRequest, artifactUploaded, err)
	}()
	artifactUploaded, err = impl.runCIStages(ciContext, ciCdRequest)
	log.Println(util.DEVTRON, artifactUploaded, err)
	var artifactUploadErr error
	if !artifactUploaded {
		cloudHelperBaseConfig := ciRequest.GetCloudHelperBaseConfig(util.BlobStorageObjectTypeArtifact)
		artifactUploaded, artifactUploadErr = helper.ZipAndUpload(cloudHelperBaseConfig, ciCdRequest.CommonWorkflowRequest.CiArtifactFileName, ciRequest.PartSize, ciRequest.ConcurrencyMultiplier)
	}

	if err != nil {
		log.Println(util.DEVTRON, err)
		return
	}

	if artifactUploadErr != nil {
		log.Println(util.DEVTRON, "error in artifact upload: ", artifactUploadErr)
		if ciCdRequest.CommonWorkflowRequest.IsExtRun {
			log.Println(util.DEVTRON, "Ignoring artifactUploadErr")
			return
		}
		return
	}

	// sync cache
	uploadCache := func() error {
		log.Println(util.DEVTRON, " cache-push")
		err = helper.SyncCache(ciRequest)
		if err != nil {
			log.Println(err)
			if ciCdRequest.CommonWorkflowRequest.IsExtRun {
				log.Println(util.DEVTRON, "Ignoring cache upload")
				// not returning error as we are ignoring the cache upload, todo: re confirm this
				return nil
			}
			return err
		}
		log.Println(util.DEVTRON, " /cache-push")
		return nil
	}
	err = util.ExecuteWithStageInfoLog(util.PUSH_CACHE, uploadCache)
	if err != nil {
		log.Println("error in cache push", err)
	}
	return
}

// TODO: take as tech debt and break this function into parts for better code readability
func (impl *CiStage) runCIStages(ciContext cicxt.CiContext, ciCdRequest *helper.CiCdTriggerEvent) (artifactUploaded bool, err error) {

	metrics := &helper.CIMetrics{}
	start := time.Now()
	metrics.TotalStartTime = start
	artifactUploaded = false
	// change the current working directory to '/'
	err = os.Chdir(util.HOMEDIR)
	if err != nil {
		return artifactUploaded, err
	}

	// using stat to get check if WORKINGDIR exist or not
	if _, err := os.Stat(util.WORKINGDIR); os.IsNotExist(err) {
		// Creating the WORKINGDIR if in case in doesn't exit
		_ = os.Mkdir(util.WORKINGDIR, os.ModeDir)
	}

	// change the current working directory to WORKINGDIR
	err = os.Chdir(util.WORKINGDIR)
	if err != nil {
		return artifactUploaded, err
	}
	ciBuildConfi := ciCdRequest.CommonWorkflowRequest.CiBuildConfig
	buildSkipEnabled := ciBuildConfi != nil && ciBuildConfi.CiBuildType == helper.BUILD_SKIP_BUILD_TYPE
	skipCheckout := ciBuildConfi != nil && ciBuildConfi.PipelineType == helper.CI_JOB
	err = impl.prepareStep(ciCdRequest, metrics, skipCheckout)
	if err != nil {
		return artifactUploaded, err
	}

	extraEnvVars, err := impl.AddExtraEnvVariableFromRuntimeParamsToCiCdEvent(ciCdRequest.CommonWorkflowRequest)
	if err != nil {
		return artifactUploaded, err
	}
	ciCdRequest.CommonWorkflowRequest.RuntimeEnvironmentVariables = extraEnvVars
	scriptEnvs, err := util2.GetGlobalEnvVariables(ciCdRequest)
	if err != nil {
		log.Println(util.DEVTRON, "error while getting global envs", err)
		return artifactUploaded, err
	}
	// Get devtron-ci yaml
	yamlLocation := ciCdRequest.CommonWorkflowRequest.CheckoutPath
	log.Println(util.DEVTRON, "devtron-ci yaml location ", yamlLocation)
	taskYaml, err := helper.GetTaskYaml(yamlLocation)
	if err != nil {
		return artifactUploaded, err
	}
	ciCdRequest.CommonWorkflowRequest.TaskYaml = taskYaml
	if ciBuildConfi != nil && ciBuildConfi.CiBuildType == helper.MANAGED_DOCKERFILE_BUILD_TYPE {
		err = makeDockerfile(ciBuildConfi.DockerBuildConfig, ciCdRequest.CommonWorkflowRequest.CheckoutPath)
		if err != nil {
			return artifactUploaded, err
		}
	}

	refStageMap := make(map[int][]*helper.StepObject)
	for _, ref := range ciCdRequest.CommonWorkflowRequest.RefPlugins {
		refStageMap[ref.Id] = ref.Steps
	}

	var preCiStageOutVariable map[int]map[string]*commonBean.VariableObject
	start = time.Now()
	metrics.PreCiStartTime = start
	var resultsFromPlugin json.RawMessage
	if len(ciCdRequest.CommonWorkflowRequest.PreCiSteps) > 0 {
		resultsFromPlugin, preCiStageOutVariable, err = impl.runPreCiSteps(ciCdRequest, metrics, buildSkipEnabled, refStageMap, scriptEnvs, artifactUploaded)
		if err != nil {
			return artifactUploaded, err
		}
	}
	var dest string
	var digest string
	if !buildSkipEnabled {
		dest, digest, err = impl.getImageDestAndDigest(ciCdRequest, metrics, scriptEnvs, refStageMap, preCiStageOutVariable, artifactUploaded)
		if err != nil {
			return artifactUploaded, err
		}
	}
	// setting digest in global env
	helper2.SetKeyValueInGlobalSystemEnv(scriptEnvs, bean2.DigestGlobalEnvKey, digest)
	var postCiDuration float64
	start = time.Now()
	metrics.PostCiStartTime = start
	var pluginArtifacts *helper.PluginArtifacts
	if len(ciCdRequest.CommonWorkflowRequest.PostCiSteps) > 0 {
		pluginArtifacts, resultsFromPlugin, err = impl.runPostCiSteps(ciCdRequest, scriptEnvs, refStageMap, preCiStageOutVariable, metrics, artifactUploaded, dest, digest)
		postCiDuration = time.Since(start).Seconds()
		if err != nil {
			return artifactUploaded, err
		}
	}
	metrics.PostCiDuration = postCiDuration
	log.Println(util.DEVTRON, " /docker-push")

	log.Println(util.DEVTRON, " artifact-upload")
	cloudHelperBaseConfig := ciCdRequest.CommonWorkflowRequest.GetCloudHelperBaseConfig(util.BlobStorageObjectTypeArtifact)
	var artifactUploadErr error
	artifactUploaded, artifactUploadErr = helper.ZipAndUpload(cloudHelperBaseConfig, ciCdRequest.CommonWorkflowRequest.CiArtifactFileName, ciCdRequest.CommonWorkflowRequest.PartSize, ciCdRequest.CommonWorkflowRequest.ConcurrencyMultiplier)
	if artifactUploadErr != nil {
		return artifactUploaded, nil
	}
	log.Println(util.DEVTRON, " /artifact-upload")

	dest, err = impl.dockerHelper.GetDestForNatsEvent(ciCdRequest.CommonWorkflowRequest, dest)
	if err != nil {
		return artifactUploaded, err
	}
	if scriptEnvs.RuntimeEnv[bean2.ExternalCiArtifact] != "" {
		runtimeImage, runtimeDigest, err := impl.handleRuntimeParametersForCiJob(scriptEnvs.RuntimeEnv, ciCdRequest)
		if err != nil {
			log.Println(util.DEVTRON, "error in handling runtime parameters for ci job and getting runtime image and digest")
			return artifactUploaded, err
		}
		if len(runtimeImage) > 0 {
			dest = runtimeImage
			digest = runtimeDigest
		}
	}

	// scan only if ci scan enabled
	if helper.IsEventTypeEligibleToScanImage(ciCdRequest.Type) &&
		ciCdRequest.CommonWorkflowRequest.ScanEnabled {
		err = impl.runImageScanning(adaptor2.GetImageScannerExecutorBean(ciCdRequest, scriptEnvs, refStageMap, metrics, artifactUploaded, dest, digest))
		if err != nil {
			return artifactUploaded, err
		}
	}

	log.Println(util.DEVTRON, " event")
	metrics.TotalDuration = time.Since(metrics.TotalStartTime).Seconds()

	event := adaptor.NewCiCompleteEvent(ciCdRequest.CommonWorkflowRequest).WithMetrics(*metrics).
		WithDockerImage(dest).WithDigest(digest).WithIsArtifactUploaded(artifactUploaded).
		WithImageDetailsFromCR(resultsFromPlugin).WithPluginArtifacts(pluginArtifacts).
		WithTargetPlatforms(GetTargetPlatformFromCiBuildConfig(ciCdRequest.CommonWorkflowRequest.CiBuildConfig))
	err = helper.SendCiCompleteEvent(ciCdRequest.CommonWorkflowRequest, event)
	if err != nil {
		log.Println(err)
		return artifactUploaded, err
	}
	log.Println(util.DEVTRON, " /event")

	err = impl.dockerHelper.StopDocker(ciContext)
	if err != nil {
		log.Println("err", err)
		return artifactUploaded, err
	}
	return artifactUploaded, nil
}

func (impl *CiStage) runPreCiSteps(ciCdRequest *helper.CiCdTriggerEvent, metrics *helper.CIMetrics,
	buildSkipEnabled bool, refStageMap map[int][]*helper.StepObject,
	scriptEnvs *util2.ScriptEnvVariables, artifactUploaded bool) (json.RawMessage, map[int]map[string]*commonBean.VariableObject, error) {
	start := time.Now()
	metrics.PreCiStartTime = start
	if !buildSkipEnabled {
		log.Println("running PRE-CI steps")
	}
	// run pre artifact processing
	_, preCiStageOutVariable, step, err := impl.stageExecutorManager.RunCiCdSteps(helper.STEP_TYPE_PRE, ciCdRequest.CommonWorkflowRequest, ciCdRequest.CommonWorkflowRequest.PreCiSteps, refStageMap, scriptEnvs, nil, true)
	preCiDuration := time.Since(start).Seconds()
	if err != nil {
		log.Println("error in running pre Ci Steps", "err", err)
		return nil, nil, helper.NewCiStageError(err).
			WithMetrics(metrics).
			WithFailureMessage(fmt.Sprintf(workFlow.PreCiFailed.String(), step.Name)).
			WithArtifactUploaded(artifactUploaded)
	}
	scriptEnvs = scriptEnvs.ResetExistingScriptEnv()
	// considering pull images from Container repo Plugin in Pre ci steps only.
	// making it non-blocking if results are not available (in case of err)
	resultsFromPlugin, fileErr := extractOutResultsIfExists()
	if fileErr != nil {
		log.Println("error in getting results", "err", fileErr.Error())
	}
	metrics.PreCiDuration = preCiDuration
	return resultsFromPlugin, preCiStageOutVariable, nil
}

func (impl *CiStage) runBuildArtifact(ciCdRequest *helper.CiCdTriggerEvent, metrics *helper.CIMetrics,
	refStageMap map[int][]*helper.StepObject, scriptEnvs *util2.ScriptEnvVariables, artifactUploaded bool,
	preCiStageOutVariable map[int]map[string]*commonBean.VariableObject) (string, error) {
	// build
	start := time.Now()
	metrics.BuildStartTime = start
	dest, err := impl.dockerHelper.BuildArtifact(ciCdRequest.CommonWorkflowRequest) // TODO make it skipable
	metrics.BuildDuration = time.Since(start).Seconds()
	if err != nil {
		log.Println("Error in building artifact", "err", err)
		// code-block starts : run post-ci which are enabled to run on ci fail
		postCiStepsToTriggerOnCiFail := getPostCiStepToRunOnCiFail(ciCdRequest.CommonWorkflowRequest.PostCiSteps)
		if len(postCiStepsToTriggerOnCiFail) > 0 {
			log.Println("Running POST-CI steps which are enabled to RUN even on CI FAIL")
			// build success will always be false
			scriptEnvs.SystemEnv[util.ENV_VARIABLE_BUILD_SUCCESS] = "false"
			// run post artifact processing
			impl.stageExecutorManager.RunCiCdSteps(helper.STEP_TYPE_POST, ciCdRequest.CommonWorkflowRequest, postCiStepsToTriggerOnCiFail, refStageMap, scriptEnvs, preCiStageOutVariable, true)
			scriptEnvs = scriptEnvs.ResetExistingScriptEnv()
		}
		// code-block ends
		err = helper.NewCiStageError(err).
			WithMetrics(metrics).
			WithFailureMessage(workFlow.BuildFailed.String()).
			WithArtifactUploaded(artifactUploaded)
	}
	log.Println(util.DEVTRON, " Build artifact completed", "dest", dest, "err", err)
	return dest, err
}

func (impl *CiStage) extractDigest(ciCdRequest *helper.CiCdTriggerEvent, dest string, metrics *helper.CIMetrics, artifactUploaded bool) (string, error) {

	var digest string
	var err error

	extractDigestStage := func() error {
		ciBuildConfi := ciCdRequest.CommonWorkflowRequest.CiBuildConfig
		isBuildX := ciBuildConfi != nil && ciBuildConfi.DockerBuildConfig != nil && ciBuildConfi.DockerBuildConfig.CheckForBuildX()
		if isBuildX {
			digest, err = impl.dockerHelper.ExtractDigestForBuildx(dest, ciCdRequest.CommonWorkflowRequest)
		} else {
			// push to dest
			log.Println(util.DEVTRON, "Docker push Artifact", "dest", dest)
			err = impl.pushArtifact(ciCdRequest, dest, digest, metrics, artifactUploaded)
			if err != nil {
				return err
			}
			digest, err = impl.dockerHelper.ExtractDigestForBuildx(dest, ciCdRequest.CommonWorkflowRequest)
		}
		return err
	}

	err = util.ExecuteWithStageInfoLog(util.DOCKER_PUSH_AND_EXTRACT_IMAGE_DIGEST, extractDigestStage)
	return digest, err
}

func (impl *CiStage) runPostCiSteps(ciCdRequest *helper.CiCdTriggerEvent, scriptEnvs *util2.ScriptEnvVariables, refStageMap map[int][]*helper.StepObject, preCiStageOutVariable map[int]map[string]*commonBean.VariableObject, metrics *helper.CIMetrics, artifactUploaded bool, dest string, digest string) (*helper.PluginArtifacts, json.RawMessage, error) {
	log.Println("running POST-CI steps")
	// sending build success as true always as post-ci triggers only if ci gets success
	scriptEnvs.SystemEnv[util.ENV_VARIABLE_BUILD_SUCCESS] = "true"
	scriptEnvs.SystemEnv["DEST"] = dest
	scriptEnvs.SystemEnv["DIGEST"] = digest
	// run post artifact processing
	pluginArtifactsFromFile, _, step, err := impl.stageExecutorManager.RunCiCdSteps(helper.STEP_TYPE_POST, ciCdRequest.CommonWorkflowRequest, ciCdRequest.CommonWorkflowRequest.PostCiSteps, refStageMap, scriptEnvs, preCiStageOutVariable, true)
	if err != nil {
		log.Println("error in running Post Ci Steps", "err", err)
		return nil, nil, helper.NewCiStageError(err).
			WithMetrics(metrics).
			WithFailureMessage(fmt.Sprintf(workFlow.PostCiFailed.String(), step.Name)).
			WithArtifactUploaded(artifactUploaded)
	}
	scriptEnvs = scriptEnvs.ResetExistingScriptEnv()
	//sent by orchestrator if copy container image v2 is configured

	// considering pull images from Container repo Plugin in post ci steps also.
	// making it non-blocking if results are not available (in case of err)
	resultsFromPlugin, fileErr := extractOutResultsIfExists()
	if fileErr != nil {
		log.Println("error in getting results", "err", fileErr.Error())
	}
	return pluginArtifactsFromFile, resultsFromPlugin, nil
}

func (impl *CiStage) runImageScanning(imageScannerExecutor *bean2.ImageScanningExecutorBean) error {
	ciCdRequest := imageScannerExecutor.CiCdRequest
	dest, digest := imageScannerExecutor.Dest, imageScannerExecutor.Digest
	metrics, artifactUploaded := imageScannerExecutor.Metrics, imageScannerExecutor.ArtifactUploaded
	scriptEnvs, refStageMap := imageScannerExecutor.ScriptEnvs, imageScannerExecutor.RefStageMap

	imageScanningStage := func() error {
		log.Println("Image Scanning Started for digest", digest)
		scanEvent := adaptor2.GetImageScanEvent(dest, digest, ciCdRequest.CommonWorkflowRequest)
		err := helper.ExecuteImageScanningViaRest(scanEvent)
		if err != nil {
			log.Println("error in running Image Scan", "err", err)
			return helper.NewCiStageError(err).
				WithMetrics(metrics).
				WithFailureMessage(workFlow.ScanFailed.String()).
				WithArtifactUploaded(artifactUploaded)
		}
		log.Println("Image scanning completed with scanEvent", scanEvent)
		return nil
	}
	imageScanningTaskExecution := func() error {
		log.Println("Image Scanning Started")
		for _, allSteps := range ciCdRequest.CommonWorkflowRequest.ImageScanningSteps {
			scanToolId := allSteps.ScanToolId
			tasks := allSteps.Steps
			//setting scan tool id in script env
			scriptEnvs.SystemEnv[bean2.ScanToolIdGlobalEnvKey] = strconv.Itoa(scanToolId)
			// run image scanning steps
			_, _, _, err := impl.stageExecutorManager.RunCiCdSteps(helper.STEP_TYPE_SCANNING, ciCdRequest.CommonWorkflowRequest, tasks, refStageMap, scriptEnvs, nil, true)
			if err != nil {
				log.Println("error in running pre Ci Steps", "err", err)
				return helper.NewCiStageError(err).
					WithMetrics(metrics).
					WithFailureMessage(workFlow.ScanFailed.String()).
					WithArtifactUploaded(artifactUploaded)
			}
		}
		//unset scan tool id in script env
		delete(scriptEnvs.SystemEnv, bean2.ScanToolIdGlobalEnvKey)
		log.Println("Image scanning completed")
		return nil
	}
	if ciCdRequest.CommonWorkflowRequest.ExecuteImageScanningVia.IsScanMediumExternal() {
		return util.ExecuteWithStageInfoLog(util.IMAGE_SCAN, imageScanningTaskExecution)
	}
	return util.ExecuteWithStageInfoLog(util.IMAGE_SCAN, imageScanningStage)
}

func (impl *CiStage) getImageDestAndDigest(ciCdRequest *helper.CiCdTriggerEvent, metrics *helper.CIMetrics, scriptEnvs *util2.ScriptEnvVariables, refStageMap map[int][]*helper.StepObject, preCiStageOutVariable map[int]map[string]*commonBean.VariableObject, artifactUploaded bool) (string, string, error) {
	dest, err := impl.runBuildArtifact(ciCdRequest, metrics, refStageMap, scriptEnvs, artifactUploaded, preCiStageOutVariable)
	if err != nil {
		return "", "", err
	}
	digest, err := impl.extractDigest(ciCdRequest, dest, metrics, artifactUploaded)
	if err != nil {
		log.Println("Error in extracting digest", "err", err)
		return "", "", err
	}
	return dest, digest, nil
}

func getPostCiStepToRunOnCiFail(postCiSteps []*helper.StepObject) []*helper.StepObject {
	var postCiStepsToTriggerOnCiFail []*helper.StepObject
	if len(postCiSteps) > 0 {
		for _, postCiStep := range postCiSteps {
			if postCiStep.TriggerIfParentStageFail {
				postCiStepsToTriggerOnCiFail = append(postCiStepsToTriggerOnCiFail, postCiStep)
			}
		}
	}
	return postCiStepsToTriggerOnCiFail
}

// extractOutResultsIfExists will return the json.RawMessage results from file
// if file doesn't exist, returns nil
func extractOutResultsIfExists() (json.RawMessage, error) {
	exists, err := util.CheckFileExists(util.ResultsDirInCIRunnerPath)
	if err != nil || !exists {
		log.Println("err", err)
		return nil, err
	}
	file, err := ioutil.ReadFile(util.ResultsDirInCIRunnerPath)
	if err != nil {
		log.Println("error in reading file", "err", err.Error())
		return nil, err
	}
	return file, nil
}

func makeDockerfile(config *helper.DockerBuildConfig, checkoutPath string) error {
	dockerfilePath := helper.GetSelfManagedDockerfilePath(checkoutPath)
	dockerfileContent := config.DockerfileContent
	f, err := os.Create(dockerfilePath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(dockerfileContent)
	return err
}

func sendCIFailureEvent(ciRequest *helper.CommonWorkflowRequest, err *helper.CiStageError) {
	event := adaptor.NewCiCompleteEvent(ciRequest).
		WithMetrics(err.GetMetrics()).
		WithIsArtifactUploaded(err.IsArtifactUploaded()).
		WithFailureReason(err.GetFailureMessage())
	e := helper.SendCiCompleteEvent(ciRequest, event)
	if e != nil {
		log.Println(e)
	}
}

func (impl *CiStage) pushArtifact(ciCdRequest *helper.CiCdTriggerEvent, dest string, digest string, metrics *helper.CIMetrics, artifactUploaded bool) error {
	imageRetryCountValue := ciCdRequest.CommonWorkflowRequest.ImageRetryCount
	imageRetryIntervalValue := ciCdRequest.CommonWorkflowRequest.ImageRetryInterval
	var err error
	for i := 0; i < imageRetryCountValue+1; i++ {
		if i != 0 {
			time.Sleep(time.Duration(imageRetryIntervalValue) * time.Second)
		}
		ciContext := cicxt.BuildCiContext(context.Background(), ciCdRequest.CommonWorkflowRequest.EnableSecretMasking)
		err = impl.dockerHelper.PushArtifact(ciContext, dest)
		if err == nil {
			break
		}
		log.Println("Error in pushing artifact", "err", err)
	}
	if err != nil {
		return helper.NewCiStageError(err).
			WithMetrics(metrics).
			WithFailureMessage(workFlow.PushFailed.String()).
			WithArtifactUploaded(artifactUploaded)
	}
	return err
}

func (impl *CiStage) AddExtraEnvVariableFromRuntimeParamsToCiCdEvent(ciRequest *helper.CommonWorkflowRequest) (map[string]string, error) {
	if len(ciRequest.RuntimeEnvironmentVariables["externalCiArtifact"]) > 0 {
		var err error
		image := ciRequest.RuntimeEnvironmentVariables["externalCiArtifact"]
		if !strings.Contains(image, ":") {
			//check for tag name
			if utils.IsValidDockerTagName(image) {
				ciRequest.DockerImageTag = image
				image, err = helper.BuildDockerImagePath(ciRequest)
				if err != nil {
					log.Println("Error in building docker image", "err", err)
					return nil, err
				}
				ciRequest.RuntimeEnvironmentVariables["externalCiArtifact"] = image
			} else {
				return nil, errors.New("external-ci-artifact image is neither a url nor a tag name")
			}

		}
		if ciRequest.ShouldPullDigest {
			var useAppDockerConfigForPrivateRegistries bool
			var err error
			useAppDockerConfig, ok := ciRequest.RuntimeEnvironmentVariables["useAppDockerConfig"]
			if ok && len(useAppDockerConfig) > 0 {
				useAppDockerConfigForPrivateRegistries, err = strconv.ParseBool(useAppDockerConfig)
				if err != nil {
					fmt.Println(fmt.Sprintf("Error in parsing useAppDockerConfig runtime param to bool from string useAppDockerConfigForPrivateRegistries:- %s, err:", useAppDockerConfig), err)
					return ciRequest.RuntimeEnvironmentVariables, err
				}
			}
			var dockerAuthConfig *bean.DockerAuthConfig
			if useAppDockerConfigForPrivateRegistries {
				dockerAuthConfig = impl.dockerHelper.GetDockerAuthConfigForPrivateRegistries(ciRequest)
			}
			log.Println("image scanning plugin configured and digest not provided hence pulling image digest")
			startTime := time.Now()
			//user has not provided imageDigest in that case fetch from docker.
			imgDigest, err := impl.dockerHelper.ExtractDigestFromImage(image, ciRequest.UseDockerApiToGetDigest, dockerAuthConfig)
			if err != nil {
				fmt.Println(fmt.Sprintf("Error in extracting digest from image %s, err:", image), err)
				return ciRequest.RuntimeEnvironmentVariables, err
			}
			log.Println(fmt.Sprintf("time since extract digest from image process:- %s", time.Since(startTime).String()))
			log.Println(fmt.Sprintf("image:- %s , image digest:- %s", image, imgDigest))
			ciRequest.RuntimeEnvironmentVariables["imageDigest"] = imgDigest
		}
	}
	return ciRequest.RuntimeEnvironmentVariables, nil
}

// When externalCiArtifact is provided (run time Env at time of build) then this image will be used further in the pipeline
// imageDigest and ciProjectDetails are optional fields
func (impl *CiStage) handleRuntimeParametersForCiJob(runtimeEnv map[string]string, ciCdRequest *helper.CiCdTriggerEvent) (string, string, error) {
	log.Println(util.DEVTRON, "external ci artifact found! exiting now with success event")
	dest := runtimeEnv[bean2.ExternalCiArtifact]
	digest := runtimeEnv[bean2.ImageDigest]
	var err error
	if len(digest) == 0 {
		digest, err = impl.extractDigestForCiJob(ciCdRequest.CommonWorkflowRequest, dest)
		if err != nil {
			log.Println(util.DEVTRON, " extract digest for ci job error", "dest", dest, "err", err)
			return dest, digest, err
		}
	}
	var tempDetails []*helper.CiProjectDetailsMin
	err = json.Unmarshal([]byte(runtimeEnv[bean2.CiProjectDetails]), &tempDetails)
	if err != nil {
		fmt.Println("Error unmarshalling ciProjectDetails JSON:", err)
		fmt.Println("ignoring the error and continuing without saving ciProjectDetails")
	}
	if len(tempDetails) > 0 && len(ciCdRequest.CommonWorkflowRequest.CiProjectDetails) > 0 {
		detail := tempDetails[0]
		ciCdRequest.CommonWorkflowRequest.CiProjectDetails[0].CommitHash = detail.CommitHash
		ciCdRequest.CommonWorkflowRequest.CiProjectDetails[0].Message = detail.Message
		ciCdRequest.CommonWorkflowRequest.CiProjectDetails[0].Author = detail.Author
		ciCdRequest.CommonWorkflowRequest.CiProjectDetails[0].CommitTime = detail.CommitTime
	}
	return dest, digest, nil
}

func (impl *CiStage) extractDigestForCiJob(workflowRequest *helper.CommonWorkflowRequest, image string) (string, error) {
	var useAppDockerConfigForPrivateRegistries bool
	var err error
	useAppDockerConfig, ok := workflowRequest.RuntimeEnvironmentVariables[bean2.UseAppDockerConfig]
	if ok && len(useAppDockerConfig) > 0 {
		useAppDockerConfigForPrivateRegistries, err = strconv.ParseBool(useAppDockerConfig)
		if err != nil {
			fmt.Println(fmt.Sprintf("Error in parsing useAppDockerConfig runtime param to bool from string useAppDockerConfigForPrivateRegistries:- %s, err:", useAppDockerConfig), err)
			// would use default val of useAppDockerConfigForPrivateRegistries i.e false in case error arises
		}
	}
	var dockerAuthConfig *bean.DockerAuthConfig
	if useAppDockerConfigForPrivateRegistries {
		dockerAuthConfig = impl.dockerHelper.GetDockerAuthConfigForPrivateRegistries(workflowRequest)
	}
	startTime := time.Now()
	//user has not provided imageDigest in that case fetch from docker.
	imgDigest, err := impl.dockerHelper.ExtractDigestFromImage(image, workflowRequest.UseDockerApiToGetDigest, dockerAuthConfig)
	if err != nil {
		fmt.Println(fmt.Sprintf("Error in extracting digest from image %s, err:", image), err)
		return "", err
	}
	log.Println(fmt.Sprintf("time since extract digest from image process:- %s", time.Since(startTime).String()))
	return imgDigest, nil
}

func GetTargetPlatformFromCiBuildConfig(ciBuildConfig *helper.CiBuildConfigBean) []string {
	if ciBuildConfig == nil {
		return []string{}
	} else if ciBuildConfig.DockerBuildConfig == nil {
		return []string{}
	} else {
		return utils.ConvertTargetPlatformStringToList(ciBuildConfig.DockerBuildConfig.TargetPlatform)
	}
}

func pullCache(metrics *helper.CIMetrics, ciCdRequest *helper.CiCdTriggerEvent) error {
	// Get ci cache TODO
	log.Println(util.DEVTRON, " cache-pull")
	start := time.Now()
	metrics.CacheDownStartTime = start

	defer func() {
		log.Println(util.DEVTRON, " /cache-pull")
		metrics.CacheDownDuration = time.Since(start).Seconds()
	}()

	err := helper.GetCache(ciCdRequest.CommonWorkflowRequest)
	if err != nil {
		return err
	}
	return nil
}

func (impl *CiStage) prepareStep(ciCdRequest *helper.CiCdTriggerEvent, metrics *helper.CIMetrics, skipCheckout bool) error {
	prepareStep := func() error {
		var useBuildx bool
		if ciCdRequest.Type == util.JOBEVENT || (ciCdRequest.CommonWorkflowRequest.CiBuildConfig != nil && ciCdRequest.CommonWorkflowRequest.CiBuildConfig.PipelineType == helper.CI_JOB) {
			// in these cases we don't get docker build config, so setting useBuildx as false explicitly
			useBuildx = false
		} else {
			if ciCdRequest.CommonWorkflowRequest.CiBuildConfig != nil &&
				ciCdRequest.CommonWorkflowRequest.CiBuildConfig.DockerBuildConfig != nil {
				useBuildx = ciCdRequest.CommonWorkflowRequest.CiBuildConfig.DockerBuildConfig.CheckForBuildX()
			}
		}
		start := time.Now()
		eg := new(errgroup.Group)

		if useBuildx {
			//we can run all stages in parallel

			//first parallel stage - pull cache
			eg.Go(func() error {
				err := pullCache(metrics, ciCdRequest)
				if err != nil {
					log.Println("cache pull fails, continuing with other stages", "err", err)
				}
				//intentionally not returning error here as we want to continue with other stages even if cache pull fails
				return nil
			})

			//second parallel stage - git clone
			eg.Go(func() error {
				return gitCloneStep(ciCdRequest, impl, skipCheckout)
			})

			//third parallel stage - docker start
			eg.Go(func() error {
				log.Println(util.DEVTRON, " docker-build")
				return impl.dockerHelper.StartDockerDaemonAndDockerLogin(ciCdRequest.CommonWorkflowRequest, true)
			})

			if err := eg.Wait(); err != nil {
				log.Println("Error in executing initial steps", "err", err)
				return err
			}

		} else {
			//first run pull cache git clone in parallel and docker start in sequence

			//first parallel stage - pull cache
			eg.Go(func() error {
				err := pullCache(metrics, ciCdRequest)
				if err != nil {
					log.Println("cache pull fails, continuing with other stages", "err", err)
				}
				//intentionally not returning error here as we want to continue with other stages even if cache pull fails
				return nil
			})

			//second parallel stage - git clone
			eg.Go(func() error {
				return gitCloneStep(ciCdRequest, impl, skipCheckout)
			})

			if err := eg.Wait(); err != nil {
				log.Println("Error in executing initial steps", "err", err)
				return err
			}

			//third stage in sequence - docker start
			log.Println(util.DEVTRON, " docker-build")
			return impl.dockerHelper.StartDockerDaemonAndDockerLogin(ciCdRequest.CommonWorkflowRequest, true)

		}

		log.Println("total time for cache, git and docker start", time.Since(start).Seconds())
		return nil
	}

	err := util.ExecuteWithStageInfoLog(util.PREPARE_STEP, prepareStep)
	if err != nil {
		log.Println(util.DEVTRON, "error in PREPARE_STEP ", "err : ", err)
	}
	return err

}

func gitCloneStep(ciCdRequest *helper.CiCdTriggerEvent, impl *CiStage, skipCheckout bool) error {
	// git handling
	log.Println(util.DEVTRON, " git")
	if !skipCheckout {
		err := impl.gitManager.CloneAndCheckout(ciCdRequest.CommonWorkflowRequest.CiProjectDetails, true)
		if err != nil {
			log.Println(util.DEVTRON, "clone err", err)
			return err
		}
	}

	log.Println(util.DEVTRON, " /git")
	return nil
}
