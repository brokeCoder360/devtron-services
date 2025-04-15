// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/devtron-labs/common-lib/monitoring"
	"github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/kubewatch/api/router"
	"github.com/devtron-labs/kubewatch/pkg/asyncProvider"
	"github.com/devtron-labs/kubewatch/pkg/cluster"
	"github.com/devtron-labs/kubewatch/pkg/config"
	"github.com/devtron-labs/kubewatch/pkg/informer"
	"github.com/devtron-labs/kubewatch/pkg/informer/cluster"
	"github.com/devtron-labs/kubewatch/pkg/informer/cluster/argoCD"
	argoWf2 "github.com/devtron-labs/kubewatch/pkg/informer/cluster/argoWf/cd"
	"github.com/devtron-labs/kubewatch/pkg/informer/cluster/argoWf/ci"
	"github.com/devtron-labs/kubewatch/pkg/informer/cluster/systemExec"
	"github.com/devtron-labs/kubewatch/pkg/logger"
	"github.com/devtron-labs/kubewatch/pkg/pubsub"
	"github.com/devtron-labs/kubewatch/pkg/resource"
	"github.com/devtron-labs/kubewatch/pkg/sql"
	"github.com/devtron-labs/kubewatch/pkg/utils"
)

// Injectors from Wire.go:

func InitializeApp() (*App, error) {
	sugaredLogger := logger.NewSugaredLogger()
	monitoringRouter := monitoring.NewMonitoringRouter(sugaredLogger)
	routerImpl := api.NewRouter(sugaredLogger, monitoringRouter)
	appConfig, err := config.GetAppConfig()
	if err != nil {
		return nil, err
	}
	sqlConfig, err := sql.GetConfig()
	if err != nil {
		return nil, err
	}
	db, err := sql.NewDbConnection(appConfig, sqlConfig, sugaredLogger)
	if err != nil {
		return nil, err
	}
	customK8sHttpTransportConfig := k8s.NewCustomK8sHttpTransportConfig()
	restConfig, err := utils.GetDefaultK8sConfig()
	if err != nil {
		return nil, err
	}
	k8sUtilImpl := utils.NewK8sUtilImpl(sugaredLogger, customK8sHttpTransportConfig, restConfig)
	clusterRepositoryImpl := repository.NewClusterRepositoryImpl(db, sugaredLogger)
	pubSubClientServiceImpl, err := pubsub.NewPubSubClientServiceImpl(sugaredLogger, appConfig)
	if err != nil {
		return nil, err
	}
	informerClientImpl := resource.NewInformerClientImpl(sugaredLogger, pubSubClientServiceImpl, appConfig, k8sUtilImpl)
	runnable := asyncProvider.NewAsyncRunnable(sugaredLogger)
	informerImpl := argoCD.NewInformerImpl(sugaredLogger, appConfig, k8sUtilImpl, informerClientImpl, runnable)
	argoWfInformerImpl := argoWf.NewInformerImpl(sugaredLogger, appConfig, k8sUtilImpl, informerClientImpl, runnable)
	informerImpl2 := argoWf2.NewInformerImpl(sugaredLogger, appConfig, k8sUtilImpl, informerClientImpl, runnable)
	systemExecInformerImpl := systemExec.NewInformerImpl(sugaredLogger, appConfig, k8sUtilImpl, pubSubClientServiceImpl, informerClientImpl)
	clusterInformerImpl := cluster.NewInformerImpl(sugaredLogger, appConfig, k8sUtilImpl, clusterRepositoryImpl, informerClientImpl, informerImpl, argoWfInformerImpl, informerImpl2, systemExecInformerImpl)
	runnerImpl := informer.NewRunnerImpl(sugaredLogger, appConfig, k8sUtilImpl, clusterInformerImpl)
	app := NewApp(routerImpl, sugaredLogger, appConfig, db, runnerImpl, runnable)
	return app, nil
}
