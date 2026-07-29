package veris

import (
	"context"
	"encoding/base64"
	"html"
	"regexp"
	"sync"
)

var (
	akamaiWebV3ScriptPathRegex   = regexp.MustCompile(`type="text/javascript"\s+src="([A-Za-z0-9/_-]+)">`)
	akamaiWebSbsdScriptPathRegex = regexp.MustCompile(`src="((?:\/[A-Za-z0-9_-]+){5,}\?v=[a-z0-9-;&=]+)"`)
)

// Origin is for example "https://www.example.com"
func ExtractAkamaiWebV3URL(origin, body string) (string, bool) {
	match := akamaiWebV3ScriptPathRegex.FindStringSubmatch(body)
	if match == nil {
		return "", false
	}

	path := match[1]
	path = html.UnescapeString(path)

	return origin + path, true
}

func ExtractAkamaiWebSBSDURL(origin, body string) (string, bool) {
	match := akamaiWebSbsdScriptPathRegex.FindStringSubmatch(body)
	if match == nil {
		return "", false
	}

	path := match[1]
	path = html.UnescapeString(path)

	return origin + path, true
}

type SBSDHint int

const (
	SBSDHintNone SBSDHint = iota
	SBSDHintChallenge
	SBSDHintNonChallenge
)

type AkamaiWebGuidance struct {
	SBSDHint      SBSDHint
	SBSDV         string
	SBSDT         int
	SBSDPostURL   string
	SBSDScriptURL string

	V3ScriptURL string
}

// Automatically detect presence of SBSD challenge and parse SBSD/V3 params
// Useful to handle cases where website doesn't always return SBSD challenge
func ExtractAkamaiWebGuidance(origin, body string) (AkamaiWebGuidance, bool) {
	var guidance AkamaiWebGuidance
	var ok bool

	guidance.V3ScriptURL, ok = ExtractAkamaiWebV3URL(origin, body)
	if !ok {
		return guidance, false
	}

	guidance.SBSDScriptURL, ok = ExtractAkamaiWebSBSDURL(origin, body)
	if ok {

	}

	return guidance, true
}

// AkamaiWebV3Session maintains the state required for a coherent
// sequence of Akamai Web V3 sensors.
type AkamaiWebV3Session struct {
	client *Client
	mu     sync.Mutex

	userAgent   string
	scriptURL   string
	scriptBytes []byte
	language    string
	ip          string

	session string
}

type akamaiWebV3SessionBuilder struct {
	client *Client

	userAgent   string
	scriptURL   string
	scriptBytes []byte
	language    string
	ip          string
}

// AkamaiWebV3Session returns a builder for a new Akamai Web V3 session. It
// does not make an API request and therefore does not accept a context. Complete
// the builder with Create().
func (c *Client) AkamaiWebV3Session(userAgent, scriptURL string, scriptBytes []byte, language string) akamaiWebV3SessionBuilder {
	return akamaiWebV3SessionBuilder{
		client:      c,
		userAgent:   userAgent,
		scriptURL:   scriptURL,
		scriptBytes: scriptBytes,
		language:    language,
	}
}

// WithIP supplies the proxy IPv4
func (b akamaiWebV3SessionBuilder) WithIP(ip string) akamaiWebV3SessionBuilder {
	b.ip = ip
	return b
}

func (b akamaiWebV3SessionBuilder) Create() *AkamaiWebV3Session {
	return &AkamaiWebV3Session{
		client:      b.client,
		userAgent:   b.userAgent,
		scriptURL:   b.scriptURL,
		scriptBytes: b.scriptBytes,
		language:    b.language,
		ip:          b.ip,
	}
}

// Sensor() makes one charged Akamai Web V3 sensor-generation request.
// The script is sent only while initializing the first remote session.
func (s *AkamaiWebV3Session) Sensor(ctx context.Context, pageURL, abck, bmsz string) (string, ReportData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	request := struct {
		PageURL      string `json:"pageUrl"`
		ScriptURL    string `json:"scriptUrl"`
		ScriptBase64 string `json:"scriptBase64,omitempty"`
		BMSZ         string `json:"bmsz"`
		ABCK         string `json:"abck"`
		Language     string `json:"language"`
		UserAgent    string `json:"userAgent"`
		IP           string `json:"ip,omitempty"`
		Session      string `json:"session,omitempty"`
	}{
		PageURL:   pageURL,
		ScriptURL: s.scriptURL,
		BMSZ:      bmsz,
		ABCK:      abck,
		Language:  s.language,
		UserAgent: s.userAgent,
		IP:        s.ip,
		Session:   s.session,
	}

	if s.session == "" {
		request.ScriptBase64 = base64.StdEncoding.EncodeToString(s.scriptBytes)
	}

	var response struct {
		Sensor     string     `json:"sensor"`
		Session    string     `json:"session"`
		ReportData ReportData `json:"reportData"`
	}

	if err := s.client.post(ctx, akamaiWebV3SensorEndpoint, nil, request, &response); err != nil {
		return "", "", err
	}

	s.session = response.Session

	return response.Sensor, response.ReportData, nil
}
