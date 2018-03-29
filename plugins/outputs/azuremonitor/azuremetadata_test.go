package azuremonitor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetMetadata(t *testing.T) {
	azureMetadata := &AzureInstanceMetadata{}
	metadata, err := azureMetadata.GetInstanceMetadata()
	if err != nil {
		t.Logf("could not get metadata: %v\n", err)
	} else {
		t.Logf("resource id  \n%s", metadata.AzureResourceID)
		t.Logf("metadata is \n%v", metadata)
	}

	//fmt.Printf("metadata is \n%v", metadata)
}

func TestGetTOKEN(t *testing.T) {
	azureMetadata := &AzureInstanceMetadata{}
	token, err := azureMetadata.GetMsiToken("", "")

	require.NoError(t, err)
	t.Logf("token is %v\n", token)
	t.Logf("expiry time is %s\n", token.ExpiresAt().Format(time.RFC3339))
	t.Logf("expiry duration is %s\n", token.ExpiresInDuration().String())

	require.NotEmpty(t, token.AccessToken)
}
