package wire

import (
	"github.com/risingwavelabs/anclax/pkg/config"
	"github.com/risingwavelabs/anclax/pkg/globalctx"
	"github.com/risingwavelabs/anclax/pkg/taskcore/worker"
	"github.com/risingwavelabs/anclax/pkg/zcore/model"
)

func NewConfiguredWorker(globalCtx *globalctx.GlobalContext, cfg *config.Config, m model.ModelInterface, taskHandler worker.TaskHandler) (worker.WorkerInterface, error) {
	return worker.NewWorker(globalCtx, cfg, m, taskHandler)
}
