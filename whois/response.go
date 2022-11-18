package whois

type DomainResponse struct {
	Code           string `json:"code"`
	DomainName     string `josn:"domainName"`
	Message        string `json:"message,omitempty"`
	ExpirationDate string `json:"expirationDate,omitempty"`
}

type Response struct {
	Body []byte
}
