//go:build integration

package service

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/risingwavelabs/anclax/pkg/app/closer"
	"github.com/risingwavelabs/anclax/pkg/config"
	"github.com/risingwavelabs/anclax/pkg/hooks"
	"github.com/risingwavelabs/anclax/pkg/zcore/model"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// ANCLAX_TEST_DSN must point to a disposable database; NewModel applies migrations.
func TestConcurrentUserCreationAndRestore(t *testing.T) {
	dsn := os.Getenv("ANCLAX_TEST_DSN")
	if dsn == "" {
		t.Skip("set ANCLAX_TEST_DSN to a disposable PostgreSQL database")
	}
	ctx := context.Background()
	cfg := &config.Config{}
	cfg.Pg.DSN = &dsn
	lib := &config.LibConfig{Pg: &config.PgCfg{}}
	lib.Pg.MaxConnections = 8
	cm := closer.NewCloserManager()
	defer cm.Close()
	m, err := model.NewModel(cfg, lib, cm)
	require.NoError(t, err)
	db, err := pgx.Connect(ctx, dsn)
	require.NoError(t, err)
	defer db.Close(ctx)
	ctrl := gomock.NewController(t)
	h := hooks.NewMockAnclaxHookInterface(ctrl)
	h.EXPECT().OnOrgCreated(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	h.EXPECT().OnUserCreated(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
	svc := &Service{m: m, hooks: h, generateSaltAndHash: func(string) (string, string, error) { return "salt", "hash", nil }}
	var before int
	require.NoError(t, db.QueryRow(ctx, "SELECT count(*) FROM anclax.orgs").Scan(&before))
	start := make(chan struct{})
	errs := make([]error, 2)
	users := make([]*UserMeta, 2)
	var wg sync.WaitGroup
	for i := range errs {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			users[i], errs[i] = svc.CreateNewUser(ctx, "concurrent-user", "password")
		}(i)
	}
	close(start)
	wg.Wait()
	var winner *UserMeta
	for i, err := range errs {
		if err == nil {
			require.Nil(t, winner)
			winner = users[i]
		} else {
			require.ErrorIs(t, err, ErrUsernameExists)
		}
	}
	require.NotNil(t, winner)
	var count int
	require.NoError(t, db.QueryRow(ctx, "SELECT count(*) FROM anclax.users WHERE name='concurrent-user'").Scan(&count))
	require.Equal(t, 1, count)
	require.NoError(t, db.QueryRow(ctx, "SELECT count(*) FROM anclax.orgs").Scan(&count))
	require.Equal(t, before+1, count, "losing transaction must roll back its org")
	require.ErrorIs(t, svc.RestoreUserByName(ctx, "concurrent-user"), ErrUsernameExists, "active user must never be restored")
	require.NoError(t, svc.DeleteUserByName(ctx, "concurrent-user"))
	_, err = svc.CreateNewUser(ctx, "concurrent-user", "password")
	require.ErrorIs(t, err, ErrUsernameExists, "deleted names remain reserved")
	start = make(chan struct{})
	for i := range errs {
		wg.Add(1)
		go func(i int) { defer wg.Done(); <-start; errs[i] = svc.RestoreUserByName(ctx, "concurrent-user") }(i)
	}
	close(start)
	wg.Wait()
	successes := 0
	for _, err := range errs {
		if err == nil {
			successes++
		} else {
			require.ErrorIs(t, err, ErrUsernameExists)
		}
	}
	require.Equal(t, 1, successes)
	restored, err := svc.GetUserByUserName(ctx, "concurrent-user")
	require.NoError(t, err)
	require.Equal(t, winner, restored)
}
