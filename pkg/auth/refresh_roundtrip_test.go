package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/risingwavelabs/anclax/pkg/config"
	"github.com/risingwavelabs/anclax/pkg/hooks"
	"github.com/risingwavelabs/anclax/pkg/macaroons"
	"github.com/risingwavelabs/anclax/pkg/macaroons/store"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// Exercise the actual signed wire format, rather than a mocked Parse result.
func TestRefreshTokenRoundTrip(t *testing.T) {
	ctrl := gomock.NewController(t)
	keys := store.NewMockKeyStore(ctrl)
	keys.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
	var signingKey []byte
	keys.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, key []byte, _ time.Duration, _ *int32) (int64, error) {
		signingKey = key
		return 2, nil
	})
	keys.EXPECT().Get(gomock.Any(), int64(2)).DoAndReturn(func(context.Context, int64) ([]byte, error) { return signingKey, nil }).AnyTimes()
	parser := macaroons.NewCaveatParser()
	require.NoError(t, parser.Register("path", func() macaroons.Caveat { return &pathCaveat{} }))
	manager := macaroons.NewMacaroonManager(keys, parser)
	hook := hooks.NewMockAnclaxHookInterface(ctrl)
	hook.EXPECT().OnUserTokensCreated(gomock.Any(), int32(7), gomock.Any()).Return(nil)
	a, err := NewAuth(&config.Config{}, manager, parser, hook)
	require.NoError(t, err)
	_, refresh, err := a.CreateUserTokens(context.Background(), 7, 11, &pathCaveat{Typ: "path", Path: "/allowed"})
	require.NoError(t, err)
	_, decoded, err := a.ParseRefreshToken(context.Background(), refresh.StringToken())
	require.NoError(t, err)
	require.Equal(t, int32(7), *decoded.UserID)
	require.Equal(t, []macaroons.Caveat{&pathCaveat{Typ: "path", Path: "/allowed"}, NewUserContextCaveat(7, 11)}, decoded.AccessToken.Caveats)
	// The payload uses the legacy object format and retains application caveats.
	wire, err := json.Marshal(decoded)
	require.NoError(t, err)
	require.Contains(t, string(wire), `"access_token":{"caveats":[`)
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error { return decoded.AccessToken.Caveats[0].Validate(c) })
	response, err := app.Test(httptest.NewRequest("GET", "/denied", nil))
	require.NoError(t, err)
	require.NotEqual(t, 200, response.StatusCode)
	require.NoError(t, response.Body.Close())
	parts := strings.Split(refresh.StringToken(), ".")
	parts[1] = base64.StdEncoding.EncodeToString([]byte(`{"type":"refresh_only"}`))
	_, _, err = a.ParseRefreshToken(context.Background(), strings.Join(parts, "."))
	require.ErrorIs(t, err, macaroons.ErrInvalidSignature)

}

type pathCaveat struct {
	Typ  string `json:"type"`
	Path string `json:"path"`
}

func (p *pathCaveat) Type() string { return p.Typ }
func (p *pathCaveat) Validate(c *fiber.Ctx) error {
	if c.Path() != p.Path {
		return macaroons.ErrCaveatCheckFailed
	}
	return nil
}

func TestRefreshCaveatRejectsInvalidPayload(t *testing.T) {
	parser := macaroons.NewCaveatParser()
	_, err := NewAuth(&config.Config{}, nil, parser, nil)
	require.NoError(t, err)
	for _, payload := range []string{
		`{"type":"refresh_only"}`,
		`{"type":"refresh_only","access_token":null}`,
		`{"type":"refresh_only","access_token":{"caveats":[{"type":"unknown"}]}}`,
		`{"type":"refresh_only","access_token":{"caveats":[{"type":"user_context","user_id":"bad"}]}}`,
	} {
		t.Run(payload, func(t *testing.T) {
			_, err := parser.Parse(base64.StdEncoding.EncodeToString([]byte(payload)))
			require.Error(t, err)
		})
	}
}
