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
	metrics := testutil.MockMetrics()
	metricsList, err := azmon.flattenMetrics(metrics)

	if err != nil {
		t.Logf("Error translating metrics %s", err)
	}

	jsonBytes, err := json.Marshal(&metricsList)
	req, err := azmon.postData(&jsonBytes)
	if err != nil {
		t.Logf("Error publishing metrics %s", err)
		//t.Logf("failed request is %#v\n", req)

		raw, err := httputil.DumpRequest(req, true)
		if err != nil {
			t.Logf("Request detail is \n%s\n", raw)
		}
	}
}
