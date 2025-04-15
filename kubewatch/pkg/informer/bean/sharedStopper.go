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

type SharedStopper struct {
	stopperChannel chan struct{}
}

func (s *SharedStopper) HasInformer() bool {
	return s != nil && s.stopperChannel != nil
}

func (s *SharedStopper) GetStopper(stopper chan struct{}) *SharedStopper {
	if !s.HasInformer() {
		return newSharedStopper(stopper)
	}
	return s
}

func (s *SharedStopper) Stop() {
	if !s.HasInformer() {
		return
	}
	if s.stopperChannel != nil {
		close(s.stopperChannel)
	}
}

func newSharedStopper(stopper chan struct{}) *SharedStopper {
	return &SharedStopper{
		stopperChannel: stopper,
	}
}
