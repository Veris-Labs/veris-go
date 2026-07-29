package veris

import (
	"context"
	"encoding/base64"
	"sync"
)

// AkamaiWebSBSDSession maintains the opaque state required for a coherent
// sequence of Akamai Web SBSD sensors. Calls on one session are serialized;
// separate sessions can be used concurrently.
type AkamaiWebSBSDSession struct {
	client *Client
	mu     sync.Mutex

	userAgent   string
	scriptURL   string
	scriptBytes []byte
	language    string

	session string
}

type akamaiWebSBSDSessionBuilder struct {
	client *Client

	userAgent   string
	scriptURL   string
	scriptBytes []byte
	language    string
}

// AkamaiWebSBSDSession returns a builder for a new Akamai Web SBSD session.
// Complete the builder with Create().
func (c *Client) AkamaiWebSBSDSession(userAgent, scriptURL string, scriptBytes []byte, language string) akamaiWebSBSDSessionBuilder {
	return akamaiWebSBSDSessionBuilder{
		client:      c,
		userAgent:   userAgent,
		scriptURL:   scriptURL,
		scriptBytes: scriptBytes,
		language:    language,
	}
}

func (b akamaiWebSBSDSessionBuilder) Create() *AkamaiWebSBSDSession {
	return &AkamaiWebSBSDSession{
		client:      b.client,
		userAgent:   b.userAgent,
		scriptURL:   b.scriptURL,
		scriptBytes: b.scriptBytes,
		language:    b.language,
	}
}

// Sensor makes one charged Akamai Web SBSD sensor-generation request.
func (s *AkamaiWebSBSDSession) Sensor(ctx context.Context, pageURL, cookie string) (string, ReportData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	request := struct {
		PageURL      string `json:"pageUrl"`
		ScriptURL    string `json:"scriptUrl"`
		ScriptBase64 string `json:"scriptBase64,omitempty"`
		Cookie       string `json:"cookie"`
		Language     string `json:"language"`
		UserAgent    string `json:"userAgent"`
		Session      string `json:"session,omitempty"`
	}{
		PageURL:      pageURL,
		ScriptURL:    s.scriptURL,
		ScriptBase64: base64.StdEncoding.EncodeToString(s.scriptBytes),
		Cookie:       cookie,
		Language:     s.language,
		UserAgent:    s.userAgent,
		Session:      s.session,
	}

	if s.session == "" {
		request.ScriptBase64 = base64.StdEncoding.EncodeToString(s.scriptBytes)
	}

	var response struct {
		Sensor     string     `json:"sensor"`
		Session    string     `json:"session"`
		ReportData ReportData `json:"reportData"`
	}

	if err := s.client.post(ctx, akamaiWebSBSDSensorEndpoint, nil, request, &response); err != nil {
		return "", "", err
	}

	s.session = response.Session

	return response.Sensor, response.ReportData, nil
}
