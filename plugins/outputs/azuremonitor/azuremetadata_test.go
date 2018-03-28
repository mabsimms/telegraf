package azuremonitor

import (
	"testing"
)

func TestGetMetadata(t *testing.T) {
	azureMetadata := &AzureInstanceMetadata{}
	metadata, err := azureMetadata.GetInstanceMetadata()
	if err != nil {
		t.Logf("could not get metadata")
	}

	//fmt.Printf("metadata is \n%v", metadata)
	t.Logf("raw metadata is \n%v", metadata.Raw)
}
