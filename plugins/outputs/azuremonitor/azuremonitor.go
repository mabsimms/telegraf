package azuremonitor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/influxdata/telegraf"
)

// AzureMonitor allows publishing of metrics to the Azure Monitor custom metrics service
type AzureMonitor struct {
	ResourceID       string `json:"resourceId"`
	HTTPPostTimeout  time.Duration
	region           string
	instanceMetadata *VirtualMachineMetadata
	msiToken         *MsiToken
	msiTokenClient   *MsiTokenClient
	bearerToken      string
}

var sampleConfig = `
## TODO
`

// Description provides a description of the plugin
func (s *AzureMonitor) Description() string {
	return "Configuration for Azure Monitor to send metrics to"
}

// SampleConfig provides a sample configuration for the plugin
func (s *AzureMonitor) SampleConfig() string {
	return sampleConfig
}

// Connect initializes the plugin and validates connectivity
func (s *AzureMonitor) Connect() error {
	return nil
}

// Close shuts down an any active connections
func (s *AzureMonitor) Close() error {
	// Close connection to the URL here
	return nil
}

// Write writes metrics to the remote endpoint
func (s *AzureMonitor) Write(metrics []telegraf.Metric) error {
	// Validate that we have a resource ID to be able to write against
	err := s.validateResourceID()
	if err != nil {
		return err
	}

	// Flatten metrics into an Azure Monitor common schema compatible format
	metricsList, err := s.flattenMetrics(metrics)
	if err != nil {
		log.Printf("Error translating metrics %s", err)
		return err
	}

	jsonBytes, err := json.Marshal(&metricsList)
	err = s.postData(&jsonBytes)
	if err != nil {
		log.Printf("Error publishing metrics %s", err)
		return err
	}

	return nil
}

func (s *AzureMonitor) validateCredentials() error {
	// TODO - config for MSI
	if true {
		if s.msiTokenClient == nil {
			s.msiTokenClient = &MsiTokenClient{}
		}

		// No token, acquire an MSI token
		if s.msiToken == nil {
			msiToken, err := s.msiTokenClient.GetMsiToken()
			if err != nil {
				return err
			}
			s.bearerToken = msiToken.AccessToken
		} else if true {
			// TODO - check for and refresh token
		}
	}

	// Otherwise check for environmental variables with AD claims and validate
	// TODO

	return nil
}

func (s *AzureMonitor) validateResourceID() error {
	// Was the resource ID manually set?
	if s.ResourceID != "" {
		return nil
	}

	// If the resource ID is empty attempt to use the instance
	// metadata service
	if s.instanceMetadata == nil {
		metadataClient := &AzureInstanceMetadata{}
		metadata, err := metadataClient.GetInstanceMetadata()
		if err != nil {
			// TODO - cannot retrieve metadata and no resource ID specified
			return err
		}

		s.ResourceID = metadata.AzureResourceID
	}
	return nil
}

type azureMonitorMetric struct {
	Time time.Time        `json:"time"`
	Data azureMonitorData `json:"data"`
}

type azureMonitorData struct {
	BaseData azureMonitorBaseData `json:"baseData"`
}

type azureMonitorBaseData struct {
	Metric         string               `json:"metric"`
	Namespace      string               `json:"namespace"`
	DimensionNames []string             `json:"dimNames"`
	Series         []azureMonitorSeries `json:"series"`
}

type azureMonitorSeries struct {
	DimensionValues []string `json:"dimValues"`
	Min             string   `json:"min"`
	Max             string   `json:"max"`
	Sum             string   `json:"sum"`
	Count           string   `json:"count"`
}

func (s *AzureMonitor) flattenMetrics(metrics []telegraf.Metric) ([]azureMonitorMetric, error) {
	var azureMetrics []azureMonitorMetric
	for _, metric := range metrics {

		// Get the list of custom dimensions (elevated tags and fields)
		var dimensionNames []string
		var dimensionValues []string
		for name, value := range metric.Fields() {
			dimensionNames = append(dimensionNames, name)
			dimensionValues = append(dimensionValues, s.formatField(value))
		}

		series := azureMonitorSeries{
			DimensionValues: dimensionValues,
		}

		if v, ok := metric.Fields()["min"]; ok {
			series.Min = s.formatField(v)
		}

		if v, ok := metric.Fields()["max"]; ok {
			series.Max = s.formatField(v)
		}

		if v, ok := metric.Fields()["sum"]; ok {
			series.Sum = s.formatField(v)
		}

		if v, ok := metric.Fields()["count"]; ok {
			series.Count = s.formatField(v)
		}

		azureMetric := azureMonitorMetric{
			Time: metric.Time(),
			Data: azureMonitorData{
				BaseData: azureMonitorBaseData{
					Metric:         metric.Name(),
					Namespace:      "default",
					DimensionNames: dimensionNames,
				},
			},
		}

		azureMetrics = append(azureMetrics, azureMetric)
	}
	return azureMetrics, nil
}

func (s *AzureMonitor) formatField(value interface{}) string {
	var ret string

	switch v := value.(type) {
	case int:
		ret = strconv.FormatInt(int64(value.(int)), 10)
	case int8:
		ret = strconv.FormatInt(int64(value.(int8)), 10)
	case int16:
		ret = strconv.FormatInt(int64(value.(int16)), 10)
	case int32:
		ret = strconv.FormatInt(int64(value.(int32)), 10)
	case int64:
		ret = strconv.FormatInt(value.(int64), 10)
	case float32:
		ret = strconv.FormatFloat(float64(value.(float32)), 'f', -1, 64)
	case float64:
		ret = strconv.FormatFloat(value.(float64), 'f', -1, 64)
	default:
		spew.Printf("field is of unsupported value type %v\n", v)
	}
	return ret
}

func (s *AzureMonitor) postData(msg *[]byte) error {
	metricsEndpoint := fmt.Sprintf("https://%s.monitoring.azure.com/%s/metrics",
		s.region, s.ResourceID)

	req, err := http.NewRequest("POST", metricsEndpoint, bytes.NewBuffer(*msg))
	if err != nil {
		log.Printf("Error creating HTTP request")
		return err
	}

	req.Header.Set("Authorization", "Bearer: "+s.bearerToken)
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{
		Timeout: s.HTTPPostTimeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		var reply []byte
		reply, err = ioutil.ReadAll(resp.Body)

		if err != nil {
			reply = nil
		}
		return fmt.Errorf("Post Error. HTTP response code:%d message:%s reply:\n%s",
			resp.StatusCode, resp.Status, reply)
	}
	return nil
}
