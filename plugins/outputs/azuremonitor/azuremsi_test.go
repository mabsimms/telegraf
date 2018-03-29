package azuremonitor

import (
	"testing"
	"time"
	"github.com/stretchr/testify/require"
)

func TestGetTOKEN(t *testing.T) {
	tokenService := &MsiTokenClient{}
	token, err := tokenService.GetMsiToken()

	require.NoError(t, err)
	t.Logf("token is %v\n", token)
	t.Logf("expiry time is %s\n", token.ExpiresAt().Format(time.RFC3339))
	t.Logf("expiry duration is %s\n", token.ExpiresInDuration().String())

	require.NotEmpty(t, token.AccessToken)
}

func TestParseTime(t *testing.T) {


}
