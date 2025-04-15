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

package application

import (
	"encoding/json"
	applicationBean "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/v2/pkg/client/clientset/versioned"
	applicationInformer "github.com/argoproj/argo-cd/v2/pkg/client/informers/externalversions/application/v1alpha1"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"log"
	"time"
)

type InformerImpl struct {
	logger *zap.SugaredLogger
	client *pubsub.PubSubClientServiceImpl
}

func NewInformerImpl(logger *zap.SugaredLogger,
	client *pubsub.PubSubClientServiceImpl) *InformerImpl {
	return &InformerImpl{
		logger: logger,
		client: client,
	}
}

func (impl *InformerImpl) GetSharedInformer(clusterId int, namespace string, k8sConfig *rest.Config) (cache.SharedIndexInformer, error) {
	startTime := time.Now()
	defer func() {
		impl.logger.Debugw("registered application informer", "namespace", namespace, "time", time.Since(startTime))
	}()
	clientSet := versioned.NewForConfigOrDie(k8sConfig)
	acdInformer := applicationInformer.NewApplicationInformer(clientSet, namespace, 0, cache.Indexers{})

	_, err := acdInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			impl.logger.Debug("app added")

			if app, ok := obj.(*applicationBean.Application); ok {
				impl.logger.Debugf("new app detected: %s, status: %s", app.Name, app.Status.Health.Status)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			impl.logger.Debug("app update detected")
			statusTime := time.Now()
			if oldApp, ok := old.(*applicationBean.Application); ok {
				if newApp, ok := new.(*applicationBean.Application); ok {
					if newApp.Status.History != nil && len(newApp.Status.History) > 0 {
						if oldApp.Status.History == nil || len(oldApp.Status.History) == 0 {
							impl.logger.Debug("new deployment detected")
							impl.sendAppUpdate(clusterId, newApp, statusTime)
						} else {
							impl.logger.Debugf("old deployment detected for update: %s, status:%s", oldApp.Name, oldApp.Status.Health.Status)
							oldRevision := oldApp.Status.Sync.Revision
							newRevision := newApp.Status.Sync.Revision
							oldStatus := string(oldApp.Status.Health.Status)
							newStatus := string(newApp.Status.Health.Status)
							newSyncStatus := string(newApp.Status.Sync.Status)
							oldSyncStatus := string(oldApp.Status.Sync.Status)
							if (oldRevision != newRevision) || (oldStatus != newStatus) || (newSyncStatus != oldSyncStatus) {
								impl.sendAppUpdate(clusterId, newApp, statusTime)
								impl.logger.Debug("send update app:" + oldApp.Name + ", oldRevision: " + oldRevision + ", newRevision:" +
									newRevision + ", oldStatus: " + oldStatus + ", newStatus: " + newStatus +
									", newSyncStatus: " + newSyncStatus + ", oldSyncStatus: " + oldSyncStatus)
							} else {
								impl.logger.Debug("skip updating app:" + oldApp.Name + ", oldRevision: " + oldRevision + ", newRevision:" +
									newRevision + ", oldStatus: " + oldStatus + ", newStatus: " + newStatus +
									", newSyncStatus: " + newSyncStatus + ", oldSyncStatus: " + oldSyncStatus)
							}
						}
					}
				} else {
					log.Println("app update detected, but skip updating, there is no new app")
				}
			} else {
				log.Println("app update detected, but skip updating, there is no old app")
			}
		},
		DeleteFunc: func(obj interface{}) {
			if app, ok := obj.(*applicationBean.Application); ok {
				statusTime := time.Now()
				impl.logger.Debugf("app delete detected: %s, status:%s", app.Name, app.Status.Health.Status)
				impl.sendAppDelete(clusterId, app, statusTime)
			}
		},
	})
	if err != nil {
		impl.logger.Errorw("error in adding acd event handler", "err", err, "namespace", namespace)
		return acdInformer, err
	}
	return acdInformer, nil
}

func (impl *InformerImpl) sendAppUpdate(clusterId int, app *applicationBean.Application, statusTime time.Time) {
	if impl.client == nil {
		log.Println("client is nil, don't send update")
		return
	}
	appDetail := applicationDetail{
		Application: app,
		StatusTime:  statusTime,
		ClusterId:   clusterId,
	}
	appJson, err := json.Marshal(appDetail)
	if err != nil {
		log.Println("marshal error on sending app update", err)
		return
	}
	log.Println("app update event for publish: ", string(appJson))
	var reqBody = appJson

	err = impl.client.Publish(pubsub.APPLICATION_STATUS_UPDATE_TOPIC, string(reqBody))
	if err != nil {
		log.Println("Error while publishing Request", err)
		return
	}
	log.Println("app update sent for app: " + app.Name)
}

func (impl *InformerImpl) sendAppDelete(clusterId int, app *applicationBean.Application, statusTime time.Time) {
	if impl.client == nil {
		log.Println("client is nil, don't send delete update")
		return
	}
	appDetail := applicationDetail{
		Application: app,
		StatusTime:  statusTime,
		ClusterId:   clusterId,
	}
	appJson, err := json.Marshal(appDetail)
	if err != nil {
		log.Println("marshal error on sending app delete update", err)
		return
	}
	log.Println("app delete event for publish: ", string(appJson))
	var reqBody = appJson
	err = impl.client.Publish(pubsub.APPLICATION_STATUS_DELETE_TOPIC, string(reqBody))
	if err != nil {
		log.Println("Error while publishing Request", err)
		return
	}
	log.Println("app update sent for app: " + app.Name)
}
