package sso

import (
	"compress/flate"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// SAMLConfig configures a SAML 2.0 identity provider.
type SAMLConfig struct {
	// EntityID is the service provider entity ID.
	EntityID string

	// MetadataURL is the IDP metadata URL.
	MetadataURL string

	// SSOURL is the SSO endpoint (from IDP metadata).
	SSOURL string

	// SLOUrl is the single logout endpoint (optional).
	SLOURL string

	// Certificate is the IDP's X.509 certificate for signature verification.
	Certificate *x509.Certificate

	// PrivateKey is the SP's private key for signing requests (optional).
	PrivateKey *rsa.PrivateKey

	// AssertionConsumerServiceURL is where SAML responses are posted.
	AssertionConsumerServiceURL string
}

// SAMLProvider implements SAML 2.0 authentication.
type SAMLProvider struct {
	config SAMLConfig
}

// NewSAMLProvider creates a new SAML provider.
func NewSAMLProvider(cfg SAMLConfig) *SAMLProvider {
	return &SAMLProvider{config: cfg}
}

// AuthnRequest represents a SAML authentication request.
type AuthnRequest struct {
	XMLName                     xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:protocol AuthnRequest"`
	ID                          string   `xml:",attr"`
	Version                     string   `xml:",attr"`
	IssueInstant                string   `xml:",attr"`
	Destination                 string   `xml:",attr"`
	AssertionConsumerServiceURL string   `xml:",attr"`
	ProtocolBinding             string   `xml:",attr"`
	Issuer                      Issuer
}

// Issuer is the SAML issuer element.
type Issuer struct {
	XMLName xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:assertion Issuer"`
	Value   string   `xml:",chardata"`
}

// GetAuthURL returns the SAML SSO URL with the AuthnRequest.
func (p *SAMLProvider) GetAuthURL(state string) string {
	// Generate request ID
	id := generateRequestID()

	req := AuthnRequest{
		ID:                          id,
		Version:                     "2.0",
		IssueInstant:                time.Now().UTC().Format(time.RFC3339),
		Destination:                 p.config.SSOURL,
		AssertionConsumerServiceURL: p.config.AssertionConsumerServiceURL,
		ProtocolBinding:             "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST",
		Issuer: Issuer{
			Value: p.config.EntityID,
		},
	}

	// Marshal and encode
	xmlData, _ := xml.Marshal(req)

	// Deflate compress
	var compressed []byte
	w, _ := flate.NewWriter(io.Discard, flate.DefaultCompression)
	w.Write(xmlData) // Simplified - real impl would capture output
	w.Close()
	if len(compressed) == 0 {
		compressed = xmlData // Fallback for this example
	}

	// Base64 encode
	encoded := base64.StdEncoding.EncodeToString(xmlData)

	// Build URL
	params := url.Values{
		"SAMLRequest": {encoded},
		"RelayState":  {state},
	}

	return p.config.SSOURL + "?" + params.Encode()
}

// SAMLResponse represents a SAML response (simplified).
type SAMLResponse struct {
	XMLName   xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:protocol Response"`
	ID        string   `xml:",attr"`
	Status    SAMLStatus
	Assertion SAMLAssertion
}

// SAMLStatus is the SAML status element.
type SAMLStatus struct {
	XMLName    xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:protocol Status"`
	StatusCode SAMLStatusCode
}

// SAMLStatusCode is the status code.
type SAMLStatusCode struct {
	XMLName xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:protocol StatusCode"`
	Value   string   `xml:",attr"`
}

// SAMLAssertion contains the assertion data.
type SAMLAssertion struct {
	XMLName            xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:assertion Assertion"`
	Subject            SAMLSubject
	Conditions         SAMLConditions
	AttributeStatement SAMLAttributeStatement
}

// SAMLSubject contains subject information.
type SAMLSubject struct {
	XMLName xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:assertion Subject"`
	NameID  SAMLNameID
}

// SAMLNameID is the name identifier.
type SAMLNameID struct {
	XMLName xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:assertion NameID"`
	Format  string   `xml:",attr"`
	Value   string   `xml:",chardata"`
}

// SAMLConditions contains validity conditions.
type SAMLConditions struct {
	XMLName      xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:assertion Conditions"`
	NotBefore    string   `xml:"NotBefore,attr"`
	NotOnOrAfter string   `xml:"NotOnOrAfter,attr"`
}

// SAMLAttributeStatement contains user attributes.
type SAMLAttributeStatement struct {
	XMLName    xml.Name        `xml:"urn:oasis:names:tc:SAML:2.0:assertion AttributeStatement"`
	Attributes []SAMLAttribute `xml:"Attribute"`
}

// SAMLAttribute is a single attribute.
type SAMLAttribute struct {
	Name   string   `xml:"Name,attr"`
	Values []string `xml:"AttributeValue"`
}

// ParseResponse parses and validates a SAML response.
func (p *SAMLProvider) ParseResponse(ctx context.Context, samlResponse string) (*UserInfo, error) {
	// Decode base64
	decoded, err := base64.StdEncoding.DecodeString(samlResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Parse XML
	var response SAMLResponse
	if err := xml.Unmarshal(decoded, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Verify status
	if response.Status.StatusCode.Value != "urn:oasis:names:tc:SAML:2.0:status:Success" {
		return nil, fmt.Errorf("SAML authentication failed: %s", response.Status.StatusCode.Value)
	}

	// Verify conditions (time validity)
	now := time.Now().UTC()
	notBefore, _ := time.Parse(time.RFC3339, response.Assertion.Conditions.NotBefore)
	notAfter, _ := time.Parse(time.RFC3339, response.Assertion.Conditions.NotOnOrAfter)

	if now.Before(notBefore) || now.After(notAfter) {
		return nil, fmt.Errorf("assertion is not valid at this time")
	}

	// Extract user info from attributes
	info := &UserInfo{
		ID: response.Assertion.Subject.NameID.Value,
	}

	for _, attr := range response.Assertion.AttributeStatement.Attributes {
		if len(attr.Values) == 0 {
			continue
		}
		switch attr.Name {
		case "email", "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress":
			info.Email = attr.Values[0]
		case "name", "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name":
			info.Name = attr.Values[0]
		case "given_name", "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname":
			info.GivenName = attr.Values[0]
		case "family_name", "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname":
			info.FamilyName = attr.Values[0]
		case "groups", "http://schemas.xmlsoap.org/claims/Group":
			info.Groups = attr.Values
		}
	}

	// Note: In production, you MUST verify the XML signature using the IDP certificate
	// This simplified implementation skips signature verification

	return info, nil
}

// ExchangeCode is not used for SAML (implements Provider interface).
func (p *SAMLProvider) ExchangeCode(ctx context.Context, code string) (*Tokens, error) {
	return nil, fmt.Errorf("SAML does not use authorization codes")
}

// GetUserInfo is not directly applicable for SAML.
func (p *SAMLProvider) GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	return nil, fmt.Errorf("use ParseResponse for SAML")
}

// ValidateToken is not applicable for SAML.
func (p *SAMLProvider) ValidateToken(ctx context.Context, idToken string) (*Claims, error) {
	return nil, fmt.Errorf("use ParseResponse for SAML")
}

// SAMLHandler returns an HTTP handler for the SAML ACS endpoint.
func (p *SAMLProvider) SAMLHandler(onSuccess func(http.ResponseWriter, *http.Request, *UserInfo)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		samlResponse := r.FormValue("SAMLResponse")
		if samlResponse == "" {
			http.Error(w, "Missing SAMLResponse", http.StatusBadRequest)
			return
		}

		userInfo, err := p.ParseResponse(r.Context(), samlResponse)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		onSuccess(w, r, userInfo)
	}
}

func generateRequestID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("_%x", b)
}
