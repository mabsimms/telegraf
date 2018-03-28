package azuremonitor

import (
	"fmt"
	"testing"
)

func TestGetMetadata(t *testing.T) {
	azureMetadata := &AzureInstanceMetadata{}
	metadata, err := azureMetadata.GetInstanceMetadata()
	if err != nil {

	}
	fmt.Printf("metadata is \n%s", metadata)
}
