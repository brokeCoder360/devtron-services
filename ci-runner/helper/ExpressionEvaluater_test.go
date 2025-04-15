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
	"fmt"
	"github.com/devtron-labs/common-lib/workflow"
	"testing"
)

// Test cases for evaluateExpression function
func Test_evaluateExpression(t *testing.T) {
	// Define a struct to hold the arguments for the test cases
	type args struct {
		condition *ConditionObject
		variables []*workflow.VariableObject
	}
	// Define a struct to hold the test case data
	type testCase struct {
		name       string
		args       args
		wantStatus bool
		wantError  error
	}
	t.Run("CASE_ERROR_HANDLING", func(t *testing.T) {
		tests := []testCase{
			{name: "UNDEFINED_VARIABLE_ERROR",
				args: args{condition: &ConditionObject{
					ConditionOnVariable: "UNDEFINED_VARIABLE",
					ConditionalOperator: "==",
					ConditionalValue:    "0",
				}, variables: []*workflow.VariableObject{}},
				wantStatus: false,
				wantError:  fmt.Errorf("variable %q not found", "UNDEFINED_VARIABLE"),
			},
			{name: "UNSUPPORTED_OPERATOR_ERROR",
				args: args{condition: &ConditionObject{
					ConditionOnVariable: "age",
					ConditionalOperator: "UNSUPPORTED_OPERATOR",
					ConditionalValue:    "10",
				}, variables: []*workflow.VariableObject{{Name: "age", Value: "10", Format: workflow.FormatTypeString}}},
				wantStatus: false,
				wantError:  fmt.Errorf("Cannot transition token types from VARIABLE [variableOperand] to VARIABLE [%s]", "UNSUPPORTED_OPERATOR"),
			},
			{name: "INVALID_VARIABLE_FORMAT_ERROR",
				args: args{condition: &ConditionObject{
					ConditionOnVariable: "age",
					ConditionalOperator: ">",
					ConditionalValue:    "10",
				}, variables: []*workflow.VariableObject{{Name: "age", Value: "NOT_A_NUMBER", Format: workflow.FormatTypeNumber}}},
				wantStatus: false,
				wantError:  fmt.Errorf("strconv.ParseFloat: parsing %q: invalid syntax", "NOT_A_NUMBER"),
			},
			{name: "INVALID_CONDITION_FORMAT_ERROR",
				args: args{condition: &ConditionObject{
					ConditionOnVariable: "age",
					ConditionalOperator: ">",
					ConditionalValue:    "NOT_A_NUMBER",
				}, variables: []*workflow.VariableObject{{Name: "age", Value: "10", Format: workflow.FormatTypeNumber}}},
				wantStatus: false,
				wantError:  fmt.Errorf("strconv.ParseFloat: parsing %q: invalid syntax", "NOT_A_NUMBER"),
			},
			{name: "INVALID_CONDITION_OBJECT_ERROR",
				args: args{condition: nil,
					variables: []*workflow.VariableObject{{Name: "age", Value: "10", Format: workflow.FormatTypeString}}},
				wantStatus: false,
				wantError:  fmt.Errorf("invalid condition object"),
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				gotStatus, err := evaluateExpression(tt.args.condition, tt.args.variables)
				if err == nil {
					t.Errorf("evaluateExpression(), wantedError = %v", tt.wantError)
				}
				if err.Error() != tt.wantError.Error() {
					t.Errorf("evaluateExpression() error = %v, wantedError %v", err, tt.wantError)
				}
				if gotStatus != tt.wantStatus {
					t.Errorf("evaluateExpression() gotStatus = %v, wantedStatus %v", gotStatus, tt.wantStatus)
				}
			})
		}
	})
	t.Run("CASE_EVAL_NUMBER", func(t *testing.T) {
		tests := []testCase{
			{name: "GREATER_THAN_EVAL_FALSE",
				args: args{condition: &ConditionObject{
					ConditionOnVariable: "age",
					ConditionalOperator: ">",
					ConditionalValue:    "10",
				}, variables: []*workflow.VariableObject{{Name: "age", Value: "8.5", Format: workflow.FormatTypeNumber}}},
				wantStatus: false,
			},
			{name: "GREATER_THAN_EVAL_TRUE",
				args: args{condition: &ConditionObject{
					ConditionOnVariable: "age",
					ConditionalOperator: ">",
					ConditionalValue:    "10.333333333333333333333333333333333333333333333",
				}, variables: []*workflow.VariableObject{{Name: "age", Value: "12", Format: workflow.FormatTypeNumber}}},
				wantStatus: true,
			},

			{name: "LESS_THAN_EVAL_FALSE",
				args: args{condition: &ConditionObject{
					ConditionOnVariable: "age",
					ConditionalOperator: "<",
					ConditionalValue:    "10",
				}, variables: []*workflow.VariableObject{{Name: "age", Value: "8.0", Format: workflow.FormatTypeNumber}}},
				wantStatus: true,
			},
			{name: "LESS_THAN_EVAL_TRUE",
				args: args{condition: &ConditionObject{
					ConditionOnVariable: "age",
					ConditionalOperator: "<",
					ConditionalValue:    "12.00000000000000000000000000000000000000000000000000005",
				}, variables: []*workflow.VariableObject{{Name: "age", Value: "10", Format: workflow.FormatTypeNumber}}},
				wantStatus: true,
			},
			{name: "EQUAL_TO_EVAL_FALSE",
				args: args{condition: &ConditionObject{
					ConditionOnVariable: "age",
					ConditionalOperator: "==",
					ConditionalValue:    "8",
				}, variables: []*workflow.VariableObject{{Name: "age", Value: "10", Format: workflow.FormatTypeNumber}}},
				wantStatus: false,
			},
			{name: "EQUAL_TO_EVAL_TRUE",
				args: args{condition: &ConditionObject{
					ConditionOnVariable: "age",
					ConditionalOperator: "==",
					ConditionalValue:    "8.00000000000000000000000000000000000000000000000000005",
				}, variables: []*workflow.VariableObject{{Name: "age", Value: "8", Format: workflow.FormatTypeNumber}}},
				wantStatus: true,
			},
			{name: "NOT_EQUAL_TO_EVAL_FALSE",
				args: args{condition: &ConditionObject{
					ConditionOnVariable: "age",
					ConditionalOperator: "!=",
					ConditionalValue:    "8.00000000000000000000000000000000000000000000000000005",
				}, variables: []*workflow.VariableObject{{Name: "age", Value: "8", Format: workflow.FormatTypeNumber}}},
				wantStatus: false,
			},
			{name: "NOT_EQUAL_TO_EVAL_TRUE",
				args: args{condition: &ConditionObject{
					ConditionOnVariable: "age",
					ConditionalOperator: "!=",
					ConditionalValue:    "11",
				}, variables: []*workflow.VariableObject{{Name: "age", Value: "10.99999", Format: workflow.FormatTypeNumber}}},
				wantStatus: true,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				gotStatus, err := evaluateExpression(tt.args.condition, tt.args.variables)
				if err != nil {
					t.Errorf("evaluateExpression(), gotError = %v", err)
				}
				if gotStatus != tt.wantStatus {
					t.Errorf("evaluateExpression() gotStatus = %v, wantedStatus %v", gotStatus, tt.wantStatus)
				}
			})
		}
	})
	t.Run("CASE_EVAL_DATE", func(t *testing.T) {
		tests := []testCase{
			{name: "GREATER_THAN_EVAL_FALSE",
				args: args{condition: &ConditionObject{
					ConditionOnVariable: "today",
					ConditionalOperator: ">",
					ConditionalValue:    "Tue Apr 12 13:55:21 IST 2022",
				}, variables: []*workflow.VariableObject{{Name: "today", Value: "Tue Apr 10 13:55:21 IST 2022", Format: workflow.FormatTypeDate}}},
				wantStatus: false,
			},
			{name: "GREATER_THAN_EVAL_TRUE",
				args: args{condition: &ConditionObject{
					ConditionOnVariable: "today",
					ConditionalOperator: ">",
					ConditionalValue:    "Tue Apr 10 13:55:21 IST 2022",
				}, variables: []*workflow.VariableObject{{Name: "today", Value: "Tue Apr 12 13:55:21 IST 2022", Format: workflow.FormatTypeDate}}},
				wantStatus: true,
			},
			{name: "LESS_THAN_EVAL_FALSE",
				args: args{condition: &ConditionObject{
					ConditionOnVariable: "today",
					ConditionalOperator: "<",
					ConditionalValue:    "Tue Apr 08 13:55:21 IST 2022",
				}, variables: []*workflow.VariableObject{{Name: "today", Value: "Tue Apr 10 13:55:21 IST 2022", Format: workflow.FormatTypeDate}}},
				wantStatus: false,
			},
			{name: "LESS_THAN_EVAL_TRUE",
				args: args{condition: &ConditionObject{
					ConditionOnVariable: "today",
					ConditionalOperator: "<",
					ConditionalValue:    "Tue Apr 12 13:55:21 IST 2022",
				}, variables: []*workflow.VariableObject{{Name: "today", Value: "Tue Apr 10 13:55:21 IST 2022", Format: workflow.FormatTypeDate}}},
				wantStatus: true,
			},
			{name: "EQUAL_TO_EVAL_FALSE",
				args: args{condition: &ConditionObject{
					ConditionOnVariable: "today",
					ConditionalOperator: "==",
					ConditionalValue:    "Tue Apr 08 13:55:21 IST 2022",
				}, variables: []*workflow.VariableObject{{Name: "today", Value: "Tue Apr 10 13:55:21 IST 2022", Format: workflow.FormatTypeDate}}},
				wantStatus: false,
			},
			{name: "EQUAL_TO_EVAL_TRUE",
				args: args{condition: &ConditionObject{
					ConditionOnVariable: "today",
					ConditionalOperator: "==",
					ConditionalValue:    "Tue Apr 08 13:55:21 IST 2022",
				}, variables: []*workflow.VariableObject{{Name: "today", Value: "Tue Apr 08 13:55:21 IST 2022", Format: workflow.FormatTypeDate}}},
				wantStatus: true,
			},
			{name: "NOT_EQUAL_TO_EVAL_FALSE",
				args: args{condition: &ConditionObject{
					ConditionOnVariable: "today",
					ConditionalOperator: "!=",
					ConditionalValue:    "Tue Apr 08 13:55:21 IST 2022",
				}, variables: []*workflow.VariableObject{{Name: "today", Value: "Tue Apr 08 13:55:21 IST 2022", Format: workflow.FormatTypeDate}}},
				wantStatus: false,
			},
			{name: "NOT_EQUAL_TO_EVAL_TRUE",
				args: args{condition: &ConditionObject{
					ConditionOnVariable: "today",
					ConditionalOperator: "!=",
					ConditionalValue:    "Tue Apr 08 13:55:21 IST 2022",
				}, variables: []*workflow.VariableObject{{Name: "today", Value: "Tue Apr 10 13:55:21 IST 2022", Format: workflow.FormatTypeDate}}},
				wantStatus: true,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				gotStatus, err := evaluateExpression(tt.args.condition, tt.args.variables)
				if err != nil {
					t.Errorf("evaluateExpression(), gotError = %v", err)
				}
				if gotStatus != tt.wantStatus {
					t.Errorf("evaluateExpression() gotStatus = %v, wantedStatus %v", gotStatus, tt.wantStatus)
				}
			})
		}
	})
	t.Run("CASE_EVAL_BOOL", func(t *testing.T) {
		tests := []testCase{
			{name: "EQUAL_TO_EVAL_FALSE",
				args: args{condition: &ConditionObject{
					ConditionOnVariable: "isValid",
					ConditionalOperator: "==",
					ConditionalValue:    "true",
				}, variables: []*workflow.VariableObject{{Name: "isValid", Value: "false", Format: workflow.FormatTypeBool}}},
				wantStatus: false,
			},
			{name: "EQUAL_TO_EVAL_TRUE",
				args: args{condition: &ConditionObject{
					ConditionOnVariable: "isValid",
					ConditionalOperator: "==",
					ConditionalValue:    "TRUE",
				}, variables: []*workflow.VariableObject{{Name: "isValid", Value: "true", Format: workflow.FormatTypeBool}}},
				wantStatus: true,
			},
			{name: "NOT_EQUAL_TO_EVAL_FALSE",
				args: args{condition: &ConditionObject{
					ConditionOnVariable: "isValid",
					ConditionalOperator: "!=",
					ConditionalValue:    "t",
				}, variables: []*workflow.VariableObject{{Name: "isValid", Value: "f", Format: workflow.FormatTypeBool}}},
				wantStatus: true,
			},
			{name: "NOT_EQUAL_TO_EVAL_TRUE",
				args: args{condition: &ConditionObject{
					ConditionOnVariable: "isValid",
					ConditionalOperator: "!=",
					ConditionalValue:    "true",
				}, variables: []*workflow.VariableObject{{Name: "isValid", Value: "FALSE", Format: workflow.FormatTypeBool}}},
				wantStatus: true,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				gotStatus, err := evaluateExpression(tt.args.condition, tt.args.variables)
				if err != nil {
					t.Errorf("evaluateExpression(), gotError = %v", err)
				}
				if gotStatus != tt.wantStatus {
					t.Errorf("evaluateExpression() gotStatus = %v, wantedStatus %v", gotStatus, tt.wantStatus)
				}
			})
		}
	})
}
