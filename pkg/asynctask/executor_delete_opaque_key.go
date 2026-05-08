package asynctask

import (
	"context"

	"github.com/risingwavelabs/anclax/core"
	"github.com/risingwavelabs/anclax/pkg/zgen/taskgen"
)

func (e *Executor) ExecuteDeleteOpaqueKey(ctx context.Context, params *taskgen.DeleteOpaqueKeyParameters) error {
	return e.model.DeleteOpaqueKey(ctx, params.KeyID)
}

func (e *Executor) OnDeleteOpaqueKeyFailed(ctx context.Context, taskID int32, params *taskgen.DeleteOpaqueKeyParameters, tx core.Tx) error {
	return nil
}
