package loganalytics

import (
	"os"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestConnectAndWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	la := &LogAnalytics{
		Workspace: os.Getenv("OMS_WORKSPACE"),
		SharedKey: os.Getenv("OMS_KEY"),
	}

	// Verify that we can connect to Log Analytics
	err := la.Connect()
	require.NoError(t, err)

	// Verify that we can write a metric to Log Analytics
	err = la.Write(testutil.MockMetrics())
	require.NoError(t, err)
}
