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

package helper

import (
	"encoding/json"
	"fmt"
	commonBean "github.com/devtron-labs/common-lib/workflow"
	"strings"
)

type RefPluginObject struct {
	Id    int           `json:"id"`
	Steps []*StepObject `json:"steps"`
}

type StepType string

func (s StepType) String() string {
	return string(s)
}

const (
	STEP_TYPE_INLINE     StepType = "INLINE"
	STEP_TYPE_REF_PLUGIN StepType = "REF_PLUGIN"
	STEP_TYPE_PRE        StepType = "PRE"
	STEP_TYPE_POST       StepType = "POST"
	STEP_TYPE_SCANNING   StepType = "SCANNING"
)

type StepObject struct {
	Name                     string                       `json:"name"`
	Index                    int                          `json:"index"`
	StepType                 string                       `json:"stepType"`     // REF_PLUGIN or INLINE
	ExecutorType             ExecutorType                 `json:"executorType"` // continer_image/ shell
	RefPluginId              int                          `json:"refPluginId"`
	Script                   string                       `json:"script"`
	InputVars                []*commonBean.VariableObject `json:"inputVars"`
	ExposedPorts             map[int]int                  `json:"exposedPorts"` //map of host:container
	OutputVars               []*commonBean.VariableObject `json:"outputVars"`
	TriggerSkipConditions    []*ConditionObject           `json:"triggerSkipConditions"`
	SuccessFailureConditions []*ConditionObject           `json:"successFailureConditions"`
	DockerImage              string                       `json:"dockerImage"`
	Command                  string                       `json:"command"`
	Args                     []string                     `json:"args"`
	CustomScriptMount        *MountPath                   `json:"customScriptMount"` // destination path - storeScriptAt
	SourceCodeMount          *MountPath                   `json:"sourceCodeMount"`   // destination path - mountCodeToContainerPath
	ExtraVolumeMounts        []*MountPath                 `json:"extraVolumeMounts"` // filePathMapping
	ArtifactPaths            []string                     `json:"artifactPaths"`
	TriggerIfParentStageFail bool                         `json:"triggerIfParentStageFail"`
}

type MountPath struct {
	SrcPath string `json:"sourcePath"`
	DstPath string `json:"destinationPath"`
}

func NewMountPath(srcPath, dstPath string) *MountPath {
	return &MountPath{
		SrcPath: sanitiseMountPath(srcPath),
		DstPath: sanitiseMountPath(dstPath),
	}
}

func sanitiseMountPath(pathString string) string {
	// trim leading and trailing single quotes, if any
	strings.Trim(pathString, "'")
	// add single quotes to sanitize the src/ dst path
	return fmt.Sprintf("'%s'", pathString)
}

// ----------
type ConditionType int

const (
	TRIGGER = iota
	SKIP
	PASS
	FAIL
)

func (d ConditionType) ValueOf(conditionType string) (ConditionType, error) {
	if conditionType == "TRIGGER" {
		return TRIGGER, nil
	} else if conditionType == "SKIP" {
		return SKIP, nil
	} else if conditionType == "PASS" {
		return PASS, nil
	} else if conditionType == "FAIL" {
		return FAIL, nil
	}
	return PASS, fmt.Errorf("invalid conditionType: %s", conditionType)
}
func (d ConditionType) String() string {
	return [...]string{"TRIGGER", "SKIP", "PASS", "FAIL"}[d]
}

func (t ConditionType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

func (t *ConditionType) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	executorType, err := t.ValueOf(s)
	if err != nil {
		return err
	}
	*t = executorType
	return nil
}

// ------

// ---------------
type ExecutorType int

const (
	CONTAINER_IMAGE ExecutorType = iota
	SHELL
	PLUGIN // Added to avoid un-marshaling error in REF_PLUGIN type steps, otherwise this value won't be used
)

func (t ExecutorType) ValueOf(executorType string) (ExecutorType, error) {
	if executorType == "CONTAINER_IMAGE" {
		return CONTAINER_IMAGE, nil
	} else if executorType == "SHELL" {
		return SHELL, nil
	} else if executorType == "PLUGIN" {
		return PLUGIN, nil
	}
	return SHELL, fmt.Errorf("invalid executorType:  %s", executorType)
}

func (t ExecutorType) String() string {
	return [...]string{"CONTAINER_IMAGE", "SHELL"}[t]
}

func (t ExecutorType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

func (t *ExecutorType) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	execType, err := t.ValueOf(s)
	if err != nil {
		return err
	}
	*t = execType
	return nil
}

// -----
