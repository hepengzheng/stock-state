//go:build wireinject
// +build wireinject

package main

import (
	"context"

	"github.com/google/wire"

	"github.com/hepengzheng/stock-state/pkg/config"
	"github.com/hepengzheng/stock-state/pkg/stockstate"
	"github.com/hepengzheng/stock-state/pkg/storage"
	"github.com/hepengzheng/stock-state/server"
)

func initApp(context.Context, *config.Config) (*server.StockServer, error) {
	panic(wire.Build(stockstate.ProviderSet, storage.ProviderSet, server.ProviderSet, providerSet))
}
