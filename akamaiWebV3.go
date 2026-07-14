package veris

import (
	"context"
	"regexp"
	"strings"
	"sync"
)

var (
	akamaiWebV3ScriptPathRegex = regexp.MustCompile(`type="text\/javascript"\s+src="([A-Za-z0-9/\-_]+)">`)
)

// Origin is for example "https://www.example.com"
func ExtractAkamaiWebV3ScriptURL(origin, body string) (string, bool) {
	matches := akamaiWebV3ScriptPathRegex.FindStringSubmatch(body)
	if len(matches) < 2 {
		return "", false
	}

	path := matches[1]

	if strings.HasPrefix(path, "/") {
		return origin + path, true
	}

	return matches[1], true
}

// AkamaiWebV3Session maintains the state required for a coherent
// sequence of Akamai Web V3 sensors.
type AkamaiWebV3Session struct {
	client *Client
	mu     sync.Mutex

	userAgent string
	scriptURL string
	script    string
	language  string
	ip        string

	state string
}

type akamaiWebV3SessionBuilder struct {
	client *Client

	userAgent string
	scriptURL string
	script    string
	language  string
	ip        string
}

// AkamaiWebV3Session returns a builder for a new Akamai Web V3 session. It
// does not make an API request and therefore does not accept a context. Complete
// the builder with Create().
func (c *Client) AkamaiWebV3Session(userAgent, scriptURL, script, language string) akamaiWebV3SessionBuilder {
	return akamaiWebV3SessionBuilder{
		client:    c,
		userAgent: userAgent,
		scriptURL: scriptURL,
		script:    script,
		language:  language,
	}
}

// WithIP supplies the proxy IPv4
func (b akamaiWebV3SessionBuilder) WithIP(ip string) akamaiWebV3SessionBuilder {
	b.ip = ip
	return b
}

func (b akamaiWebV3SessionBuilder) Create() *AkamaiWebV3Session {
	return &AkamaiWebV3Session{
		client:    b.client,
		userAgent: b.userAgent,
		scriptURL: b.scriptURL,
		script:    b.script,
		language:  b.language,
		ip:        b.ip,
	}
}

// Sensor() makes one charged Akamai Web V3 sensor-generation request.
// The script is sent only while initializing the first remote session.
func (s *AkamaiWebV3Session) Sensor(ctx context.Context, pageURL, abck, bmsz string) (string, ReportData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	request := struct {
		PageURL   string `json:"pageUrl"`
		ScriptURL string `json:"scriptUrl"`
		Script    string `json:"script,omitempty"`
		BMSZ      string `json:"bmsz"`
		ABCK      string `json:"abck"`
		Language  string `json:"language"`
		UserAgent string `json:"userAgent"`
		IP        string `json:"ip,omitempty"`
		Session   string `json:"session,omitempty"`
	}{
		PageURL:   pageURL,
		ScriptURL: s.scriptURL,
		BMSZ:      bmsz,
		ABCK:      abck,
		Language:  s.language,
		UserAgent: s.userAgent,
		IP:        s.ip,
		Session:   s.state,
	}

	if s.state == "" {
		request.Script = s.script
	}

	var response struct {
		Sensor     string     `json:"sensor"`
		Session    string     `json:"session"`
		ReportData ReportData `json:"reportData"`
	}

	if err := s.client.post(ctx, akamaiWebV3SensorEndpoint, nil, request, &response); err != nil {
		return "", "", err
	}

	s.state = response.Session

	return response.Sensor, response.ReportData, nil
}
