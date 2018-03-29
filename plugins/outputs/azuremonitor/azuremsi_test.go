package azuremonitor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetTOKEN(t *testing.T) {
	tokenService := &MsiTokenClient{}
	token, err := tokenService.GetMsiToken()

	require.NoError(t, err)
	t.Logf("token is %v\n", token)

	require.NotEmpty(t, token.AccessToken)
}
