package cluster

import (
	"github.com/devtron-labs/kubewatch/pkg/informer/cluster/argoCD"
	cdWf "github.com/devtron-labs/kubewatch/pkg/informer/cluster/argoWf/cd"
	ciWf "github.com/devtron-labs/kubewatch/pkg/informer/cluster/argoWf/ci"
	"github.com/devtron-labs/kubewatch/pkg/informer/cluster/systemExec"
	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	argoCD.NewInformerImpl,
	cdWf.NewInformerImpl,
	ciWf.NewInformerImpl,
	systemExec.NewInformerImpl,

	NewInformerImpl,
	wire.Bind(new(Informer), new(*InformerImpl)),
)
