package azuremonitor

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestDefaultConnectAndWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test with all defaults (MSI+IMS)
	azmon := &AzureMonitor{}

	// Verify that we can connect to Log Analytics
	err := azmon.Connect()
	require.NoError(t, err)

	// Verify that we can write a metric to Log Analytics
	err = azmon.Write(testutil.MockMetrics())
	require.NoError(t, err)
}
