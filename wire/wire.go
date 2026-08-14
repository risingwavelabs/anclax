//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/google/wire"
	"github.com/risingwavelabs/anclax/pkg/app"
	"github.com/risingwavelabs/anclax/pkg/app/closer"
	"github.com/risingwavelabs/anclax/pkg/asynctask"
	"github.com/risingwavelabs/anclax/pkg/auth"
	"github.com/risingwavelabs/anclax/pkg/config"
	"github.com/risingwavelabs/anclax/pkg/controller"
	"github.com/risingwavelabs/anclax/pkg/globalctx"
	"github.com/risingwavelabs/anclax/pkg/hooks"
	"github.com/risingwavelabs/anclax/pkg/macaroons"
	"github.com/risingwavelabs/anclax/pkg/macaroons/store"
	"github.com/risingwavelabs/anclax/pkg/metrics"
	"github.com/risingwavelabs/anclax/pkg/server"
	"github.com/risingwavelabs/anclax/pkg/service"
	taskctrl "github.com/risingwavelabs/anclax/pkg/taskcore/ctrl"
	taskcore "github.com/risingwavelabs/anclax/pkg/taskcore/store"
	"github.com/risingwavelabs/anclax/pkg/zcore/model"
	"github.com/risingwavelabs/anclax/pkg/zgen/taskgen"
)

func initializeApplication(
	cfg *config.Config,
	libCfg *config.LibConfig,
	serverOptions server.Options,
) (*app.Application, error) {
	wire.Build(
		app.NewDebugServer,
		app.NewApplication,
		closer.NewCloserManager,
		service.NewService,
		controller.NewController,
		controller.NewValidator,
		model.NewModel,
		server.NewServer,
		auth.NewAuth,
		macaroons.NewMacaroonManager,
		store.NewStore,
		taskcore.NewTaskStore,
		taskctrl.NewWorkerControlPlane,
		macaroons.NewCaveatParser,
		globalctx.New,
		metrics.NewMetricsServer,
		NewConfiguredWorker,
		taskgen.NewTaskHandler,
		taskgen.NewTaskRunner,
		asynctask.NewExecutor,
		hooks.NewBaseHook,
	)
	return nil, nil
}
