package wire

import (
	"testing"

	"github.com/risingwavelabs/anclax/pkg/config"
	"github.com/risingwavelabs/anclax/pkg/globalctx"
	"github.com/risingwavelabs/anclax/pkg/taskcore/worker"
	"github.com/stretchr/testify/require"
)

func TestNewConfiguredWorker_DefaultUsesWorker(t *testing.T) {
	gctx := globalctx.New()
	t.Cleanup(gctx.Cancel)

	w, err := NewConfiguredWorker(gctx, &config.Config{}, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, w)
	_, ok := w.(*worker.Worker)
	require.True(t, ok)
}
