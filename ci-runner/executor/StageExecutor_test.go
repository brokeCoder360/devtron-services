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

package executor

import (
	"github.com/devtron-labs/ci-runner/executor/util"
	commonBean "github.com/devtron-labs/common-lib/workflow"
	"reflect"
	"testing"
)

func Test_deduceVariables(t *testing.T) {
	type args struct {
		desiredVars          []*commonBean.VariableObject
		scriptEnvVars        *util.ScriptEnvVariables
		preCiStageVariable   map[int]map[string]*commonBean.VariableObject
		postCiStageVariables map[int]map[string]*commonBean.VariableObject
	}
	tests := []struct {
		name    string
		args    args
		want    []*commonBean.VariableObject
		wantErr bool
	}{
		{name: "only value type",
			args: args{
				desiredVars: []*commonBean.VariableObject{
					{Name: "age", Value: "20", VariableType: commonBean.VariableTypeValue, Format: commonBean.FormatTypeNumber},
					{Name: "name", Value: "test", VariableType: commonBean.VariableTypeValue, Format: commonBean.FormatTypeString},
					{Name: "status", Value: "true", VariableType: commonBean.VariableTypeValue, Format: commonBean.FormatTypeBool}},
				scriptEnvVars:        nil,
				preCiStageVariable:   nil,
				postCiStageVariables: nil,
			},
			wantErr: false,
			want: []*commonBean.VariableObject{
				{Name: "age", Value: "20", VariableType: commonBean.VariableTypeValue, Format: commonBean.FormatTypeNumber, TypedValue: float64(20)},
				{Name: "name", Value: "test", VariableType: commonBean.VariableTypeValue, Format: commonBean.FormatTypeString, TypedValue: "test"},
				{Name: "status", Value: "true", VariableType: commonBean.VariableTypeValue, Format: commonBean.FormatTypeBool, TypedValue: true}},
		}, {name: "from global",
			args: args{
				desiredVars: []*commonBean.VariableObject{
					{Name: "age", VariableType: commonBean.VariableTypeRefGlobal, Format: commonBean.FormatTypeNumber, ReferenceVariableName: "age"},
					{Name: "name", VariableType: commonBean.VariableTypeRefGlobal, Format: commonBean.FormatTypeString, ReferenceVariableName: "my-name"},
					{Name: "status", VariableType: commonBean.VariableTypeRefGlobal, Format: commonBean.FormatTypeBool, ReferenceVariableName: "status"}},
				scriptEnvVars: &util.ScriptEnvVariables{
					SystemEnv: map[string]string{"age": "20", "my-name": "test", "status": "true"},
				},
				preCiStageVariable:   nil,
				postCiStageVariables: nil,
			},
			wantErr: false,
			want: []*commonBean.VariableObject{
				{Name: "age", Value: "20", VariableType: commonBean.VariableTypeRefGlobal, Format: commonBean.FormatTypeNumber, TypedValue: float64(20), ReferenceVariableName: "age"},
				{Name: "name", Value: "test", VariableType: commonBean.VariableTypeRefGlobal, Format: commonBean.FormatTypeString, TypedValue: "test", ReferenceVariableName: "my-name"},
				{Name: "status", Value: "true", VariableType: commonBean.VariableTypeRefGlobal, Format: commonBean.FormatTypeBool, TypedValue: true, ReferenceVariableName: "status"}},
		}, {name: "REF_PRE_CI",
			args: args{
				desiredVars: []*commonBean.VariableObject{
					{Name: "age", VariableType: commonBean.VariableTypeRefPreCi, Format: commonBean.FormatTypeNumber, ReferenceVariableName: "age", ReferenceVariableStepIndex: 1},
					{Name: "name", VariableType: commonBean.VariableTypeRefPreCi, Format: commonBean.FormatTypeString, ReferenceVariableName: "my-name", ReferenceVariableStepIndex: 1},
					{Name: "status", VariableType: commonBean.VariableTypeRefPreCi, Format: commonBean.FormatTypeBool, ReferenceVariableName: "status", ReferenceVariableStepIndex: 1}},
				scriptEnvVars: &util.ScriptEnvVariables{
					SystemEnv: map[string]string{"age": "22", "my-name": "test1", "status": "false"},
				},
				preCiStageVariable: map[int]map[string]*commonBean.VariableObject{1: {"age": &commonBean.VariableObject{Name: "age", Value: "20"},
					"my-name": &commonBean.VariableObject{Name: "my-name", Value: "test"},
					"status":  &commonBean.VariableObject{Name: "status", Value: "true"},
				}},
				postCiStageVariables: nil,
			},
			wantErr: false,
			want: []*commonBean.VariableObject{
				{Name: "age", VariableType: commonBean.VariableTypeRefPreCi, Format: commonBean.FormatTypeNumber, ReferenceVariableName: "age", Value: "20", TypedValue: float64(20), ReferenceVariableStepIndex: 1},
				{Name: "name", VariableType: commonBean.VariableTypeRefPreCi, Format: commonBean.FormatTypeString, ReferenceVariableName: "my-name", Value: "test", TypedValue: "test", ReferenceVariableStepIndex: 1},
				{Name: "status", VariableType: commonBean.VariableTypeRefPreCi, Format: commonBean.FormatTypeBool, ReferenceVariableName: "status", Value: "true", TypedValue: true, ReferenceVariableStepIndex: 1}},
		}, {name: "VARIABLE_TYPE_REF_POST_CI",
			args: args{
				desiredVars: []*commonBean.VariableObject{
					{Name: "age", VariableType: commonBean.VariableTypeRefPostCi, Format: commonBean.FormatTypeNumber, ReferenceVariableName: "age", ReferenceVariableStepIndex: 1},
					{Name: "name", VariableType: commonBean.VariableTypeRefPostCi, Format: commonBean.FormatTypeString, ReferenceVariableName: "my-name", ReferenceVariableStepIndex: 1},
					{Name: "status", VariableType: commonBean.VariableTypeRefPostCi, Format: commonBean.FormatTypeBool, ReferenceVariableName: "status", ReferenceVariableStepIndex: 1}},
				scriptEnvVars: &util.ScriptEnvVariables{
					SystemEnv: map[string]string{"age": "22", "my-name": "test1", "status": "false"},
				},
				postCiStageVariables: map[int]map[string]*commonBean.VariableObject{1: {"age": &commonBean.VariableObject{Name: "age", Value: "20"},
					"my-name": &commonBean.VariableObject{Name: "my-name", Value: "test"},
					"status":  &commonBean.VariableObject{Name: "status", Value: "true"},
				}},
				preCiStageVariable: nil,
			},
			wantErr: false,
			want: []*commonBean.VariableObject{
				{Name: "age", VariableType: commonBean.VariableTypeRefPostCi, Format: commonBean.FormatTypeNumber, ReferenceVariableName: "age", Value: "20", TypedValue: float64(20), ReferenceVariableStepIndex: 1},
				{Name: "name", VariableType: commonBean.VariableTypeRefPostCi, Format: commonBean.FormatTypeString, ReferenceVariableName: "my-name", Value: "test", TypedValue: "test", ReferenceVariableStepIndex: 1},
				{Name: "status", VariableType: commonBean.VariableTypeRefPostCi, Format: commonBean.FormatTypeBool, ReferenceVariableName: "status", Value: "true", TypedValue: true, ReferenceVariableStepIndex: 1}},
		},

		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := deduceVariables(tt.args.desiredVars, tt.args.scriptEnvVars, tt.args.preCiStageVariable, tt.args.postCiStageVariables, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("deduceVariables() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("deduceVariables() got = %v, want %v", got, tt.want)
			}
		})
	}
}
