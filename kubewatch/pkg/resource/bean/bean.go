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

type SharedInformerType string

const (
	ApplicationResourceType SharedInformerType = "application"
	CiWorkflowResourceType  SharedInformerType = "ci/workflow"
	CdWorkflowResourceType  SharedInformerType = "cd/workflow"
)

type InformerFactoryType string

const (
	SecretInformerFactoryType InformerFactoryType = "secret"
)

type EventHandlers[T any] struct {
	AddFunc, DeleteFunc func(*T)
	UpdateFunc          func(*T, *T)
}

func (e *EventHandlers[T]) AddFuncHandler(handler func(*T)) *EventHandlers[T] {
	e.AddFunc = handler
	return e
}

func (e *EventHandlers[T]) DeleteFuncHandler(handler func(*T)) *EventHandlers[T] {
	e.DeleteFunc = handler
	return e
}

func (e *EventHandlers[T]) UpdateFuncHandler(handler func(*T, *T)) *EventHandlers[T] {
	e.UpdateFunc = handler
	return e
}

func NewEventHandlers[T any]() *EventHandlers[T] {
	return &EventHandlers[T]{}
}
