package azuremonitor

import (
	"fmt"
	"testing"
)

func TestGetTOKEN(t *testing.T) {
	tokenService := &MsiTokenClient{}
	token, err := tokenService.GetMsiToken()
	if err != nil {
		t.Logf("could not get token")
	}
	fmt.Printf("token is %v\n", token)
}
