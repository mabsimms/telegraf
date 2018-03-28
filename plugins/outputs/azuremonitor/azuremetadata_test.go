package azuremonitor

import (
	"testing"
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
