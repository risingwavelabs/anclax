package auth

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	macaroons "github.com/risingwavelabs/anclax/pkg/macaroons"
)

const (
	CaveatUserContext = "user_context"
	CaveatRefreshOnly = "refresh_only"
)

type UserContextCaveat struct {
	Typ    string `json:"type"`
	UserID int32  `json:"user_id"`
	OrgID  int32  `json:"org_id"`
}

func NewUserContextCaveat(userID int32, orgID int32) *UserContextCaveat {
	return &UserContextCaveat{
		Typ:    CaveatUserContext,
		UserID: userID,
		OrgID:  orgID,
	}
}

func (uc *UserContextCaveat) Type() string {
	return uc.Typ
}

func (uc *UserContextCaveat) Validate(ctx *fiber.Ctx) error {
	ctx.Locals(ContextKeyUserID, uc.UserID)
	ctx.Locals(ContextKeyOrgID, uc.OrgID)
	return nil
}

type RefreshOnlyCaveat struct {
	caveatParser macaroons.CaveatParserInterface
	Typ          string              `json:"type"`
	UserID       *int32              `json:"user_id,omitempty"`
	AccessToken  *macaroons.Macaroon `json:"access_token"`
}

func NewRefreshOnlyCaveat(userID *int32, accessToken *macaroons.Macaroon) *RefreshOnlyCaveat {
	return &RefreshOnlyCaveat{
		Typ:         CaveatRefreshOnly,
		UserID:      userID,
		AccessToken: accessToken,
	}
}

func (rc *RefreshOnlyCaveat) Type() string {
	return rc.Typ
}

func (rc *RefreshOnlyCaveat) Validate(ctx *fiber.Ctx) error {
	if ctx.Method() == "POST" && strings.HasSuffix(ctx.Path(), "/auth/refresh") {
		return nil
	}
	return errors.Wrapf(macaroons.ErrCaveatCheckFailed, "invalid request: %s %s, the token is for refresh only", ctx.Method(), ctx.Path())
}

// UnmarshalJSON uses the application's registered parsers for the embedded
// restrictions. The existing JSON format contains caveat objects, not token
// strings; decoding them directly into []Caveat cannot recover concrete types.
func (rc *RefreshOnlyCaveat) UnmarshalJSON(data []byte) error {
	var wire struct {
		Typ         string `json:"type"`
		UserID      *int32 `json:"user_id,omitempty"`
		AccessToken *struct {
			Caveats []json.RawMessage `json:"caveats"`
		} `json:"access_token"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	if wire.AccessToken == nil {
		return errors.Wrap(ErrInvalidRefreshToken, "missing access token")
	}
	if rc.caveatParser == nil {
		return errors.Wrap(ErrInvalidRefreshToken, "missing caveat parser")
	}
	caveats := make([]macaroons.Caveat, 0, len(wire.AccessToken.Caveats))
	for _, raw := range wire.AccessToken.Caveats {
		caveat, err := rc.caveatParser.Parse(base64.StdEncoding.EncodeToString(raw))
		if err != nil {
			return errors.Wrap(err, "invalid access token caveat")
		}
		caveats = append(caveats, caveat)
	}
	rc.Typ, rc.UserID = wire.Typ, wire.UserID
	rc.AccessToken = &macaroons.Macaroon{Caveats: caveats}
	return nil
}
