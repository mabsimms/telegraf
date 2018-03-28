package loganalytics

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"

	"strings"
	"time"
)

// LogAnalytics is the configuration structure for the output plugin
type LogAnalytics struct {
	Workspace       string `toml:"workspace"`
	SharedKey       string `toml:"sharedkey"`
	LogName         string `toml:"logname"`
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

	json, err := s.flattenMetrics(metrics)
	if err != nil {
		log.Printf("Error translating metrics %s", err)
		return err
	}

	// TODO - do we really need to convert to byte array for this?
	jsonBytes := []byte(json)
	err = s.postData(&jsonBytes, s.LogName)
	if err != nil {
		log.Printf("Error publishing metrics %s", err)
		return err
	}

	return nil
}

func (s *LogAnalytics) flattenMetrics(metrics []telegraf.Metric) (string, error) {
	for _, metric := range metrics {
		// write `metric` to the output sink here
		fmt.Printf("%s\n", metric)
	}

	return "", nil
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

	client := http.Client{
		Timeout: s.HTTPPostTimeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		return fmt.Errorf("Post Error. HTTP response code:%d message:%s", resp.StatusCode, resp.Status)
	}
	return nil
}
