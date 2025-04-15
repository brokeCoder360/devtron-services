package repository

import "github.com/google/wire"

var WireSet = wire.NewSet(
	NewClusterRepositoryImpl,
	wire.Bind(new(ClusterRepository), new(*ClusterRepositoryImpl)),
)
