package whois

import (
	"bytes"
	"fmt"
	"net/http"
)

const (
	PublicAPI = "https://dms.inet.vn/api/public/whois/v1/whois/directly"
)

func NewRequest(domain string) (*http.Request, error) {
	jsonBody := []byte(fmt.Sprintf(`{"domainName": "%s"}`, domain))

	httpReq, err := http.NewRequest(http.MethodPost, PublicAPI, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	return httpReq, nil
}
