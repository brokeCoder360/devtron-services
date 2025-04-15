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

package api

import (
	"encoding/json"
	"fmt"
	bean2 "github.com/devtron-labs/common-lib/imageScan/bean"
	"github.com/devtron-labs/image-scanner/common"
	"github.com/devtron-labs/image-scanner/pkg/clairService"
	"github.com/devtron-labs/image-scanner/pkg/grafeasService"
	"github.com/devtron-labs/image-scanner/pkg/klarService"
	"github.com/devtron-labs/image-scanner/pkg/security"
	"github.com/devtron-labs/image-scanner/pkg/sql/adaptor"
	"github.com/devtron-labs/image-scanner/pkg/sql/bean"
	"github.com/devtron-labs/image-scanner/pkg/sql/repository"
	"github.com/devtron-labs/image-scanner/pkg/user"
	"go.uber.org/zap"
	"net/http"
	"os"
	"time"
)

type RestHandler interface {
	ScanForVulnerability(w http.ResponseWriter, r *http.Request)
	ScanForVulnerabilityEvent(scanConfig *bean2.ImageScanEvent) (*common.ScanEventResponse, error)
	RegisterAndSaveScannedResult(w http.ResponseWriter, r *http.Request)
}

func NewRestHandlerImpl(logger *zap.SugaredLogger,
	grafeasService grafeasService.GrafeasService,
	userService user.UserService, imageScanService security.ImageScanService,
	klarService klarService.KlarService,
	clairService clairService.ClairService,
	imageScanConfig *security.ImageScanConfig) *RestHandlerImpl {
	return &RestHandlerImpl{
		Logger:           logger,
		grafeasService:   grafeasService,
		userService:      userService,
		ImageScanService: imageScanService,
		KlarService:      klarService,
		ClairService:     clairService,
		imageScanConfig:  imageScanConfig,
	}
}

type RestHandlerImpl struct {
	Logger           *zap.SugaredLogger
	grafeasService   grafeasService.GrafeasService
	userService      user.UserService
	ImageScanService security.ImageScanService
	KlarService      klarService.KlarService
	ClairService     clairService.ClairService
	imageScanConfig  *security.ImageScanConfig
}
type Response struct {
	Code   int         `json:"code,omitempty"`
	Status string      `json:"status,omitempty"`
	Result interface{} `json:"result,omitempty"`
	Errors []*ApiError `json:"errors,omitempty"`
}
type ApiError struct {
	HttpStatusCode    int         `json:"-"`
	Code              string      `json:"code,omitempty"`
	InternalMessage   string      `json:"internalMessage,omitempty"`
	UserMessage       interface{} `json:"userMessage,omitempty"`
	UserDetailMessage string      `json:"userDetailMessage,omitempty"`
}

type ResetRequest struct {
	AppId         int `json:"appId"`
	EnvironmentId int `json:"environmentId"`
}

func (impl *RestHandlerImpl) ScanForVulnerability(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var scanConfig bean2.ImageScanEvent
	err := decoder.Decode(&scanConfig)
	if err != nil {
		impl.Logger.Errorw("error in decode request", "error", err)
		WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	impl.Logger.Infow("imageScan event", "scanConfig", scanConfig)
	result, err := impl.ScanForVulnerabilityEvent(&scanConfig)
	if err != nil {
		WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	impl.Logger.Debugw("save", "status", result)
	WriteJsonResp(w, err, result, http.StatusOK)
}

func (impl *RestHandlerImpl) ScanForVulnerabilityEvent(scanConfig *bean2.ImageScanEvent) (*common.ScanEventResponse, error) {
	if scanConfig.UserId == 0 {
		scanConfig.UserId = 1 //setting user as system user in case of empty user data
	}
	impl.Logger.Infow("image scan req", "req", scanConfig)

	tool, err := impl.ImageScanService.GetActiveTool()
	if err != nil {
		impl.Logger.Errorw("err in image scanning", "err", err)
		return nil, err
	}
	//creating execution history
	scanEventJson, err := json.Marshal(scanConfig)
	if err != nil {
		impl.Logger.Errorw("error in marshalling scanEvent", "event", scanConfig, "err", err)
		return nil, err
	}
	executionHistoryModel := adaptor.GetImageScanExecutionHistory(scanConfig, scanEventJson, time.Now())

	executionHistory, executionHistoryDirPath, err := impl.ImageScanService.RegisterScanExecutionHistoryAndState(executionHistoryModel, tool.Id)
	if err != nil {
		impl.Logger.Errorw("service err, RegisterScanExecutionHistoryAndState", "err", err)
		return nil, err
	}
	result, err := impl.ScanImageAsPerTool(scanConfig, tool, executionHistory, executionHistoryDirPath)
	if err != nil {
		impl.Logger.Errorw("service err, ScanImageAsPerTool", "err", err)
		return nil, err
	}
	//deleting executionDirectoryPath with files as well
	err = os.RemoveAll(executionHistoryDirPath)
	if err != nil {
		impl.Logger.Errorw("error in deleting executionHistoryDirectory", "err", err)
		return nil, err
	}
	return result, nil
}

func (impl *RestHandlerImpl) ScanImageAsPerTool(scanConfig *bean2.ImageScanEvent, tool *repository.ScanToolMetadata,
	executionHistory *repository.ImageScanExecutionHistory, executionHistoryDirPath string) (*common.ScanEventResponse, error) {
	var result = &common.ScanEventResponse{}
	imageToBeScanned, err := impl.ImageScanService.GetImageToBeScannedAndFetchCliEnv(scanConfig)
	if err != nil {
		impl.Logger.Errorw("service err, GetImageToBeScanned", "err", err)
		return nil, err
	}
	scanConfig.Image = imageToBeScanned
	if tool.Name == bean.ScanToolClair && tool.Version == bean.ScanToolVersion2 {
		result, err = impl.KlarService.Process(scanConfig, executionHistory)
		if err != nil {
			impl.Logger.Errorw("err in process msg", "err", err)
			return nil, err
		}
	} else if tool.Name == bean.ScanToolClair && tool.Version == bean.ScanToolVersion4 {
		result, err = impl.ClairService.ScanImage(scanConfig, tool, executionHistory)
		if err != nil {
			impl.Logger.Errorw("err in process msg", "err", err)
			return nil, err
		}
	} else if tool.Name == bean.ScannerTypeTrivy && tool.Version == bean.ScanToolVersion1 {
		err = impl.ImageScanService.ScanImage(scanConfig, tool, executionHistory, executionHistoryDirPath)
		if err != nil {
			impl.Logger.Errorw("err in process msg", "err", err)
			return nil, err
		}
	} else {
		err = fmt.Errorf("no tool found for scanning")
		impl.Logger.Errorw("err in process msg", "err", err)
		return nil, err
	}
	return result, nil
}

func (impl *RestHandlerImpl) RegisterAndSaveScannedResult(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var scanResultPayload bean2.ScanResultPayload
	err := decoder.Decode(&scanResultPayload)
	if err != nil {
		impl.Logger.Errorw("error in decoding scanResultPayload", "error", err)
		WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	err = ValidateScanResultPayload(&scanResultPayload)
	if err != nil {
		impl.Logger.Errorw("validation failed", "error", err)
		WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	impl.Logger.Debugw("register and save scan result payload", "saveScanResultPayload", scanResultPayload)
	_, err = impl.ImageScanService.RegisterAndSaveScannedResult(&scanResultPayload)
	if err != nil {
		impl.Logger.Errorw("service err, RegisterAndSaveScannedResult", "err", err)
		WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	WriteJsonResp(w, nil, nil, http.StatusOK)
}

func ValidateScanResultPayload(scanResultPayload *bean2.ScanResultPayload) error {
	if scanResultPayload.ScanToolId == 0 {
		return fmt.Errorf("scan tool id not found: required")
	}
	if scanResultPayload.ImageScanEvent == nil {
		return fmt.Errorf("image and digest not found: required")
	}
	if scanResultPayload.ImageScanEvent != nil && len(scanResultPayload.ImageScanEvent.Image) == 0 {
		return fmt.Errorf("image not found: required")
	}
	if scanResultPayload.ImageScanEvent != nil && len(scanResultPayload.ImageScanEvent.ImageDigest) == 0 {
		return fmt.Errorf("image digest not found: required")
	}
	if len(scanResultPayload.Sbom) == 0 {
		return fmt.Errorf("sbom not found: required")
	}
	if len(scanResultPayload.SourceScanningResult) == 0 {
		return fmt.Errorf("source Scanning result not found: required")
	}
	return nil
}
