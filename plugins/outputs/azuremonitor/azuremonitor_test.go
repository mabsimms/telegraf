package azuremonitor

import (
	"encoding/json"
	"net/http/httputil"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

// func TestDefaultConnectAndWrite(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip("Skipping integration test in short mode")
// 	}

// 	// Test with all defaults (MSI+IMS)
// 	azmon := &AzureMonitor{}

// 	// Verify that we can connect to Log Analytics
// 	err := azmon.Connect()
// 	require.NoError(t, err)

// 	// Verify that we can write a metric to Log Analytics
// 	err = azmon.Write(testutil.MockMetrics())
// 	require.NoError(t, err)
// }

func TestPostData(t *testing.T) {
	azmon := &AzureMonitor{}
 	err := azmon.Connect()

	metrics := testutil.MockMetrics()
	metricsList, err := azmon.flattenMetrics(metrics)

	jsonBytes, err := json.Marshal(&metricsList)
	t.Logf("json content is:\n%s\n", string(jsonBytes))

	req, err := azmon.postData(&jsonBytes)
	if err != nil {
		t.Logf("Error publishing metrics %s", err)
		t.Logf("url is %+v\n", req.URL)
		t.Logf("failed request is %+v\n", req)

		raw, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			t.Logf("Request detail is \n%s\n", string(raw))
		} else { 
			t.Logf("could not dump request: %s\n", err)
		}
	}
}
