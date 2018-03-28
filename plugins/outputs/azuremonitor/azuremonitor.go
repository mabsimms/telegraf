package azuremonitor

import (
	"time"

	"github.com/influxdata/telegraf"
)

type AzureMonitor struct {
	ResourceId string `json:"resourceId"`
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
	return nil
}

func (s *AzureMonitor) validateResourceId() {
	// Was the resource ID manually set?
	if s.ResourceId != "" {
		return
	}

	// If the resource ID is empty attempt to use the instance
	// metadata service
	// TODO
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

	// switch v := value.(type) {
	// case int:
	// 	ret = strconv.FormatInt(value.(int), 10)
	// case int8:
	// 	ret = strconv.FormatInt(v.(int8), 10)
	// case int16:
	// 	ret = strconv.FormatInt(v.(int16), 10)
	// case int32:
	// 	ret = strconv.FormatInt(v.(int32), 10)
	// case int64:
	// 	ret = strconv.FormatInt(v.(int64), 10)
	// case float32:
	// 	ret = strconv.FormatFloat(v.(float32), 'f', -1, 64)
	// case float64:
	// 	ret = strconv.FormatFloat(v.(float64), 'f', -1, 64)
	// default:
	// 	spew.Printf("field is of unsupported value type %v\n", v)
	// }
	return ret
}

func (s *AzureMonitor) postData(msg *[]byte) {

}
