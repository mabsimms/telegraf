package azuremonitor

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/prometheus/common/log"
)

type AzureInstanceMetadata struct {
}

type VirtualMachineMetadata struct {
}

const (
	url = "http://169.254.169.254/metadata/instance?api-version=2017-08-01&format=json"
)

func (s *AzureInstanceMetadata) GetInstanceMetadata() (*VirtualMachineMetadata, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorf("Error creating HTTP request")
		return nil, err
	}
	req.Header.Set("Metadata", "true")
	client := http.Client{
		Timeout: 15,
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

	fmt.Printf("%s\n", reply)
	return nil, nil
}
