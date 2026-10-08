package service

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/risingwavelabs/anclax/pkg/macaroons/store"
	"testing"

	"github.com/pkg/errors"
	"github.com/risingwavelabs/anclax/pkg/auth"
	"github.com/risingwavelabs/anclax/pkg/hooks"
	"github.com/risingwavelabs/anclax/pkg/zcore/model"
	"github.com/risingwavelabs/anclax/pkg/zgen/querier"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreateNewUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockModel := model.NewMockModelInterfaceWithTransaction(ctrl)
	mockAuth := auth.NewMockAuthInterface(ctrl)
	mockHooks := hooks.NewMockAnclaxHookInterface(ctrl)

	var (
		orgID  = int32(101)
		userID = int32(102)
		org    = &querier.AnclaxOrg{
			ID: orgID,
		}
		user = &querier.AnclaxUser{
			ID: userID,
		}
		username = "testuser"
		password = "testpassword"
		salt     = "salt"
		hash     = "hash"
		ctx      = context.Background()
	)

	mockModel.EXPECT().CreateOrg(ctx, fmt.Sprintf("%s's Org", username)).Return(org, nil)

	mockHooks.EXPECT().OnOrgCreated(ctx, gomock.Any(), org.ID).Return(nil)

	mockHooks.EXPECT().OnUserCreated(ctx, gomock.Any(), user.ID).Return(nil)

	mockModel.EXPECT().CreateUser(ctx, querier.CreateUserParams{
		Name:         username,
		PasswordHash: hash,
		PasswordSalt: salt,
	}).Return(user, nil)

	mockModel.EXPECT().InsertOrgOwner(ctx, querier.InsertOrgOwnerParams{
		UserID: userID,
		OrgID:  orgID,
	}).Return(nil, nil)

	mockModel.EXPECT().InsertOrgUser(ctx, querier.InsertOrgUserParams{
		UserID: userID,
		OrgID:  orgID,
	}).Return(nil, nil)

	mockModel.EXPECT().SetUserDefaultOrg(ctx, querier.SetUserDefaultOrgParams{
		UserID: userID,
		OrgID:  orgID,
	}).Return(nil)

	service := &Service{
		m:     mockModel,
		hooks: mockHooks,
		auth:  mockAuth,
		generateSaltAndHash: func(inputPassword string) (string, string, error) {
			if inputPassword != password {
				return "", "", errors.New("password mismatch")
			}
			return salt, hash, nil
		},
	}

	u, err := service.CreateNewUser(ctx, username, password)
	require.NoError(t, err)
	require.Equal(t, orgID, u.OrgID)

}

func TestUpdateUserPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockModel := model.NewMockModelInterfaceWithTransaction(ctrl)
	mockAuth := auth.NewMockAuthInterface(ctrl)
	mockHooks := hooks.NewMockAnclaxHookInterface(ctrl)

	var (
		userID = int32(102)
		user   = &querier.AnclaxUser{
			ID: userID,
		}
		username = "testuser"
		password = "newpassword"
		salt     = "newsalt"
		hash     = "newhash"
		ctx      = context.Background()
	)

	mockModel.EXPECT().GetUserByName(ctx, username).Return(user, nil)

	mockModel.EXPECT().UpdateUserPassword(ctx, querier.UpdateUserPasswordParams{
		ID:           userID,
		PasswordHash: hash,
		PasswordSalt: salt,
	}).Return(nil)

	service := &Service{
		m:     mockModel,
		hooks: mockHooks,
		auth:  mockAuth,
		generateSaltAndHash: func(inputPassword string) (string, string, error) {
			if inputPassword != password {
				return "", "", errors.New("password mismatch")
			}
			return salt, hash, nil
		},
	}

	resultUserID, err := service.UpdateUserPassword(ctx, username, password)
	require.NoError(t, err)
	require.Equal(t, userID, resultUserID)
}

func TestCreateNewUserConflict(t *testing.T) {
	for _, constraint := range []string{"users_name_key", "other_constraint"} {
		t.Run(constraint, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			m := model.NewMockModelInterfaceWithTransaction(ctrl)
			h := hooks.NewMockAnclaxHookInterface(ctrl)
			m.EXPECT().CreateOrg(gomock.Any(), gomock.Any()).Return(&querier.AnclaxOrg{ID: 1}, nil)
			h.EXPECT().OnOrgCreated(gomock.Any(), gomock.Any(), int32(1)).Return(nil)
			pgErr := &pgconn.PgError{Code: "23505", ConstraintName: constraint}
			m.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(nil, pgErr)
			svc := &Service{m: m, hooks: h, generateSaltAndHash: func(string) (string, string, error) { return "salt", "hash", nil }}
			_, err := svc.CreateNewUser(context.Background(), "same", "password")
			if constraint == "users_name_key" {
				require.ErrorIs(t, err, ErrUsernameExists)
			} else {
				require.ErrorIs(t, err, pgErr)
			}
		})
	}
}

func TestRestoreUserRequiresDeletedRow(t *testing.T) {
	for _, count := range []int64{0, 1} {
		ctrl := gomock.NewController(t)
		m := model.NewMockModelInterface(ctrl)
		m.EXPECT().RestoreUserByName(gomock.Any(), "same").Return(count, nil)
		svc := &Service{m: m}
		err := svc.RestoreUserByName(context.Background(), "same")
		if count == 0 {
			require.ErrorIs(t, err, ErrUsernameExists)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestRefreshTokenRevoked(t *testing.T) {
	ctrl := gomock.NewController(t)
	a := auth.NewMockAuthInterface(ctrl)
	a.EXPECT().ParseRefreshToken(gomock.Any(), "revoked").Return(nil, nil, errors.Wrap(store.ErrKeyNotFound, "parse"))
	svc := &Service{auth: a}
	_, err := svc.RefreshToken(context.Background(), "revoked")
	require.ErrorIs(t, err, ErrRefreshTokenExpired)
}
