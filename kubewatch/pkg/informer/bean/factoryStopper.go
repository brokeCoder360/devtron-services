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

import kubeinformers "k8s.io/client-go/informers"

type FactoryStopper struct {
	stopperChannel  chan struct{}
	informerFactory kubeinformers.SharedInformerFactory
}

func (s *FactoryStopper) HasInformer() bool {
	return s != nil && s.stopperChannel != nil
}

func (s *FactoryStopper) GetStopper(informerFactory kubeinformers.SharedInformerFactory, stopper chan struct{}) *FactoryStopper {
	if !s.HasInformer() {
		return newFactoryStopper(informerFactory, stopper)
	}
	return s
}

func (s *FactoryStopper) Stop() {
	if !s.HasInformer() {
		return
	}
	if s.stopperChannel != nil {
		close(s.stopperChannel)
		// Shutdown informer factory only after, closing the stopper channel.
		// As the informer factory will be stopped only after the stopper channel is closed.
		if s.informerFactory != nil {
			s.informerFactory.Shutdown()
		}
	}
}

func newFactoryStopper(informerFactory kubeinformers.SharedInformerFactory,
	stopper chan struct{}) *FactoryStopper {
	return &FactoryStopper{
		informerFactory: informerFactory,
		stopperChannel:  stopper,
	}
}
