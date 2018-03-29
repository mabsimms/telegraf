package azuremonitor

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// MsiToken is the managed service identity token
type MsiToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    string `json:"expires_in"`
	ExpiresOn    string `json:"expires_on"`
	NotBefore    string `json:"not_before"`
	Resource     string `json:"resource"`
	TokenType    string `json:"token_type"`
}

// MsiTokenClient is the client for accessing and validating MSI tokens
type MsiTokenClient struct {
}

// GetMsiToken retrieves a managed service identity token from the specified port on the local VM
func (s *MsiTokenClient) GetMsiToken() (*MsiToken, error) {
	// Create HTTP request for MSI token to access Azure Resource Manager
	var msiEndpoint *url.URL
	msiEndpoint, err := url.Parse("http://localhost:50342/oauth2/token")
	if err != nil {
		return nil, err
	}

	msiParameters := url.Values{}
	msiParameters.Add("resource", "https://management.azure.com/")
	msiEndpoint.RawQuery = msiParameters.Encode()
	req, err := http.NewRequest("GET", msiEndpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Metadata", "true")

	// Create the HTTP client and call the token service
	client := http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// Complete reading the body
	defer resp.Body.Close()

	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		return nil, fmt.Errorf("Post Error. HTTP response code:%d message:%s",
			resp.StatusCode, resp.Status)
	}

	reply, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var token MsiToken
	if err := json.Unmarshal(reply, &token); err != nil {
		return nil, err
	}

	return &token, nil
}
