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

package pubsub

import (
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/kubewatch/pkg/config"
	"go.uber.org/zap"
)

func NewPubSubClientServiceImpl(logger *zap.SugaredLogger, appConfig *config.AppConfig) (*pubsub.PubSubClientServiceImpl, error) {
	if appConfig.GetExternalConfig().External {
		logger.Warnw("external config found, skipping pubsub client creation")
		logger.Debugw("skipping pubsub client creation", "externalConfig", appConfig.GetExternalConfig())
		return nil, nil
	}
	client, err := pubsub.NewPubSubClientServiceImpl(logger)
	if err != nil {
		logger.Errorw("error in startup", "err", err)
	}
	return client, err
}
