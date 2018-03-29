package azuremonitor

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
	"strconv"
)

// MsiToken is the managed service identity token
type MsiToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    string  `json:"expires_in"`
	ExpiresOn    string  `json:"expires_on"`
	NotBefore    string  `json:"not_before"`
	Resource     string `json:"resource"`
	TokenType    string `json:"token_type"`

	expiresAt    time.Time
	notBefore    time.Time
}

func (m *MsiToken) parseTimes() {
	val, err := strconv.ParseInt(m.ExpiresOn, 10, 64)
	if err == nil {
		m.expiresAt = time.Unix(val, 0)
	}

	val, err = strconv.ParseInt(m.NotBefore, 10, 64)
	if err == nil {
		m.notBefore = time.Unix(val, 0)
	}
}

func (m *MsiToken) ExpiresAt() time.Time { 
	return m.expiresAt
}

// ExpiresInDuration returns the duration until the token expires
func (m *MsiToken) ExpiresInDuration() time.Duration {
	expiresDuration := m.expiresAt.Sub(time.Now().UTC())
	return expiresDuration
}

// NotBeforeTime returns the time at which the token becomes valid
func (m *MsiToken) NotBeforeTime() time.Time {
	return m.notBefore
}

// MsiTokenClient is the client for accessing and validating MSI tokens
type MsiTokenClient struct {
}

// GetMsiToken retrieves a managed service identity token from the specified port on the local VM
func (s *MsiTokenClient) GetMsiToken() (*MsiToken, error) {
	// Acquire an MSI token.  Documented at:
	// https://docs.microsoft.com/en-us/azure/active-directory/managed-service-identity/how-to-use-vm-token

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
	token.parseTimes()

	return &token, nil
}
