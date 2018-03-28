package azuremonitor

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/prometheus/common/log"
)

type AzureInstanceMetadata struct {
}

// VirtualMachineMetadata contains information about a VM from the metadata service
type VirtualMachineMetadata struct {
	Raw             string
	AzureResourceID string
	Compute         struct {
		Location             string `json:"location"`
		Name                 string `json:"name"`
		Offer                string `json:"offer"`
		OsType               string `json:"osType"`
		PlacementGroupID     string `json:"placementGroupId"`
		PlatformFaultDomain  string `json:"platformFaultDomain"`
		PlatformUpdateDomain string `json:"platformUpdateDomain"`
		Publisher            string `json:"publisher"`
		ResourceGroupName    string `json:"resourceGroupName"`
		Sku                  string `json:"sku"`
		SubscriptionID       string `json:"subscriptionId"`
		Tags                 string `json:"tags"`
		Version              string `json:"version"`
		VMID                 string `json:"vmId"`
		VMScaleSetName       string `json:"vmScaleSetName"`
		VMSize               string `json:"vmSize"`
		Zone                 string `json:"zone"`
	} `json:"compute"`
	Network struct {
		Interface []struct {
			Ipv4 struct {
				IPAddress []struct {
					PrivateIPAddress string `json:"privateIpAddress"`
					PublicIPAddress  string `json:"publicIpAddress"`
				} `json:"ipAddress"`
				Subnet []struct {
					Address string `json:"address"`
					Prefix  string `json:"prefix"`
				} `json:"subnet"`
			} `json:"ipv4"`
			Ipv6 struct {
				IPAddress []interface{} `json:"ipAddress"`
			} `json:"ipv6"`
			MacAddress string `json:"macAddress"`
		} `json:"interface"`
	} `json:"network"`
}

const (
	url = "http://169.254.169.254/metadata/instance?api-version=2017-12-01"
)

func (s *AzureInstanceMetadata) GetInstanceMetadata() (*VirtualMachineMetadata, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorf("Error creating HTTP request")
		return nil, err
	}
	req.Header.Set("Metadata", "true")
	client := http.Client{
		Timeout: 15 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	reply, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		return nil, fmt.Errorf("Post Error. HTTP response code:%d message:%s reply:\n%s",
			resp.StatusCode, resp.Status, reply)
	}

	var metadata VirtualMachineMetadata
	if err := json.Unmarshal(reply, &metadata); err != nil {
		return nil, err
	}
	metadata.AzureResourceID = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/virtualMachines/%s",
		metadata.Compute.SubscriptionID, metadata.Compute.ResourceGroupName, metadata.Compute.Name)

	return &metadata, nil
}
