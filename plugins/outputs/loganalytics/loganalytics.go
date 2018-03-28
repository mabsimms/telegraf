package loganalytics

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

// LogAnalytics is the configuration structure for the output plugin
type LogAnalytics struct {
	Workspace       string   `toml:"workspace"`
	SharedKey       string   `toml:"sharedkey"`
	LogName         string   `toml:"logname"`
	IncludeTags     []string `toml:"includeTags"`
	URL             string
	HTTPPostTimeout time.Duration
}

var sampleConfig = `
## The name of the workspace for your log analytics instance
## Recommended to pull this from an environment varaible
workspace = "$OMS_WORKSPACE"
## The shared key of the workspace for your log analytics instance
## Recommended to pull this from an environment varaible
sharedKey = "$OMS_KEY"
## The set of tags to include in the output records
includeTags = [ "host", dc" ] # optional.
`

// Description provides a description of the plugin
func (s *LogAnalytics) Description() string {
	return "Configuration for Azure Log Analytics to send metrics to"
}

// SampleConfig provides a sample configuration for the plugin
func (s *LogAnalytics) SampleConfig() string {
	return sampleConfig
}

const (
	method      = "POST"
	contentType = "application/json"
	resource    = "/api/logs"
)

// Connect initializes the plugin and validates connectivity
func (s *LogAnalytics) Connect() error {
	// Validate the configuration
	if s.SharedKey == "" || s.Workspace == "" {
		return fmt.Errorf("Log analytics workspace or shared key not defined")
	}

	if s.LogName == "" {
		return fmt.Errorf("Log analytics log name not defined")
	}

	s.URL = "https://" + s.Workspace + ".ods.opinsights.azure.com" + resource + "?api-version=2016-04-01"
	s.HTTPPostTimeout = time.Second * 30

	// Make a connection to the URL here (TODO health check)
	endpointURL, err := url.Parse(s.URL)
	if err != nil {
		return fmt.Errorf("Not a valid endpoint " + s.URL)
	}

	_, err = net.LookupHost(endpointURL.Host)
	if err != nil {
		return fmt.Errorf("Could not resolve endpoint " + s.URL)
	}
	return nil
}

// Close shuts down an any active connections
func (s *LogAnalytics) Close() error {
	// Close connection to the URL here
	return nil
}

// Write writes metrics to the remote endpoint
func (s *LogAnalytics) Write(metrics []telegraf.Metric) error {
	// Flatten metrics into a log-analytics compatible format
	jsonBytes, err := s.flattenMetrics(metrics)
	if err != nil {
		log.Printf("Error translating metrics %s", err)
		return err
	}

	err = s.postData(&jsonBytes, s.LogName)
	if err != nil {
		log.Printf("Error publishing metrics %s", err)
		return err
	}

	return nil
}

func (s *LogAnalytics) flattenMetrics(metrics []telegraf.Metric) ([]byte, error) {
	var events []map[string]string

	for _, metric := range metrics {
		timestamp := metric.Time()
		instance := ""

		// Pull out the tag array
		taglist := make(map[string]string)
		for _, tagName := range s.IncludeTags {
			if val, ok := metric.Tags()[tagName]; ok {
				taglist[tagName] = val
			}
		}

		for fieldName, fieldValue := range metric.Fields() {
			val := 0.0

			switch v := fieldValue.(type) {
			case int:
				val = float64(v)
			case int8:
				val = float64(v)
			case int16:
				val = float64(v)
			case int32:
				val = float64(v)
			case int64:
				val = float64(v)
			case float32:
				val = float64(v)
			case float64:
				val = v
			default:
				spew.Printf("field is of unsupported value type %v\n", v)
			}

			event := map[string]string{
				"timestamp": timestamp.Format(time.RFC3339),
				"category":  metric.Name(),
				"instance":  instance,
				"metric":    fieldName,
				"value":     strconv.FormatFloat(val, 'f', -1, 64),
			}

			for tagName, tagValue := range taglist {
				event[tagName] = tagValue
			}

			events = append(events, event)
		}
	}

	jsonString, err := json.Marshal(events)
	if err != nil {
		fmt.Printf("Error in json converstion %s\n", err)
	}

	return jsonString, nil
}

type metricRecord struct {
	Timestamp time.Time `json:"timestamp"`
	Hostname  string    `json:"hostname"`
	Category  string    `json:"category"`
	Instance  string    `json:"instance"`
	Metric    string    `json:"metric"`
	Value     float64   `json:"value"`
}

func init() {
	outputs.Add("loganalytics", func() telegraf.Output { return &LogAnalytics{} })
}

// TODO _ change to an actual date object (to enforce rfc1123)
func (s *LogAnalytics) buildSignature(date string, contentLength int,
	method string, contentType string, resource string) (string, error) {

	xHeaders := "x-ms-date:" + date
	stringToHash := method + "\n" + strconv.Itoa(contentLength) + "\n" + contentType + "\n" + xHeaders + "\n" + resource
	bytesToHash := []byte(stringToHash)
	keyBytes, err := base64.StdEncoding.DecodeString(s.SharedKey)
	if err != nil {
		return "", err
	}
	hasher := hmac.New(sha256.New, keyBytes)
	hasher.Write(bytesToHash)
	encodedHash := base64.StdEncoding.EncodeToString(hasher.Sum(nil))
	authorization := fmt.Sprintf("SharedKey %s:%s", s.Workspace, encodedHash)
	return authorization, err
}

// TODO - add time to signature here
// PostData posts message to OMS
func (s *LogAnalytics) postData(msg *[]byte, logType string) error {

	// Headers
	contentLength := len(*msg)
	rfc1123date := time.Now().UTC().Format(time.RFC1123)

	// TODO: rfc1123 date should have UTC offset
	rfc1123date = strings.Replace(rfc1123date, "UTC", "GMT", 1)

	// Signature
	signature, err := s.buildSignature(rfc1123date, contentLength, method, contentType, resource)
	if err != nil {
		log.Printf("Error building signature")
		return err
	}
	// Create request
	req, err := http.NewRequest("POST", s.URL, bytes.NewBuffer(*msg))
	if err != nil {
		log.Printf("Error creating HTTP request")
		return err
	}

	req.Header.Set("Authorization", signature)
	req.Header.Set("Log-Type", logType)
	req.Header["x-ms-date"] = []string{rfc1123date}
	req.Header.Set("Content-Type", "application/json")

	//spew.Dump(signature)
	//spew.Dump(req)

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
