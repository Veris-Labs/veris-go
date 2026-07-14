package veris

import (
	"context"
	"net/url"
	"sync"
)

// AkamaiBMPIOSSession maintains the opaque state required for a coherent
// sequence of Akamai BMP iOS sensors. Sensor generations on one session are
// serialized; separate sessions can be used concurrently.
type AkamaiBMPIOSSession struct {
	client *Client

	bmpVersion  string
	appPackage  string
	appVersion  string
	language    string
	udid        string
	startMillis int64
	model       string
	iosVersion  string

	mu    sync.Mutex
	state string
}

// akamaiBMPIOSSessionBuilder configures optional iOS device-selection and
// sensor-generation values before a session is created.
type akamaiBMPIOSSessionBuilder struct {
	client *Client

	bmpVersion string
	appPackage string
	appVersion string
	language   string

	iosVersion    string
	minIOSVersion string
	maxIOSVersion string
	model         string
	ip            string
}

// AkamaiBMPIOSSession returns a builder for a new Akamai BMP iOS session. It
// does not make an API request and therefore does not accept a context. Complete
// the builder with Create, which makes the session-initialization request.
// Creating a session is charged at the normal sensor-generation rate even when
// the returned session is never used, so it should be created only when the caller
// intends to generate at least one sensor. First sensor is free.
func (c *Client) AkamaiBMPIOSSession(bmpVersion, appPackage, appVersion string) akamaiBMPIOSSessionBuilder {
	return akamaiBMPIOSSessionBuilder{
		client:     c,
		bmpVersion: bmpVersion,
		appPackage: appPackage,
		appVersion: appVersion,
	}
}

func (b akamaiBMPIOSSessionBuilder) WithLanguage(language string) akamaiBMPIOSSessionBuilder {
	b.language = language
	return b
}

// WithIOSVersion restricts device selection to an exact iOS version. It cannot
// be combined with WithIOSVersionRange.
func (b akamaiBMPIOSSessionBuilder) WithIOSVersion(version string) akamaiBMPIOSSessionBuilder {
	b.iosVersion = version
	return b
}

// WithIOSVersionRange restricts device selection to an inclusive iOS version
// range. An empty minimum or maximum leaves that side of the range unbounded.
// It cannot be combined with WithIOSVersion.
func (b akamaiBMPIOSSessionBuilder) WithIOSVersionRange(minimum, maximum string) akamaiBMPIOSSessionBuilder {
	b.minIOSVersion = minimum
	b.maxIOSVersion = maximum
	return b
}

// WithModel restricts device selection to the given model identifier.
// Model identifiers are typically of the form "iPhone14,2".
func (b akamaiBMPIOSSessionBuilder) WithModel(model string) akamaiBMPIOSSessionBuilder {
	b.model = model
	return b
}

func (b akamaiBMPIOSSessionBuilder) WithIP(ip string) akamaiBMPIOSSessionBuilder {
	b.ip = ip
	return b
}

// Create makes the charged API request that initializes an Akamai BMP iOS
// session. The returned session owns the selected device and all opaque sensor
// state. An initialized but unused session is still charged at the normal
// sensor-generation rate, so it should be created only when the caller
// intends to generate at least one sensor. First sensor is free.
func (b akamaiBMPIOSSessionBuilder) Create(ctx context.Context) (*AkamaiBMPIOSSession, error) {
	query := url.Values{}
	if b.iosVersion != "" {
		query.Set("iosVersion", b.iosVersion)
	}
	if b.minIOSVersion != "" {
		query.Set("minIosVersion", b.minIOSVersion)
	}
	if b.maxIOSVersion != "" {
		query.Set("maxIosVersion", b.maxIOSVersion)
	}
	if b.model != "" {
		query.Set("model", b.model)
	}
	if b.ip != "" {
		query.Set("ip", b.ip)
	}

	var response struct {
		Metadata struct {
			UDID        string `json:"udid"`
			StartMillis int64  `json:"startMillis"`
			Model       string `json:"model"`
			IOSVersion  string `json:"iosVersion"`
		} `json:"metadata"`
		Session string `json:"session"`
	}
	if err := b.client.get(ctx, akamaiBMPIOSInitEndpoint, query, &response); err != nil {
		return nil, err
	}

	return &AkamaiBMPIOSSession{
		client:      b.client,
		bmpVersion:  b.bmpVersion,
		appPackage:  b.appPackage,
		appVersion:  b.appVersion,
		language:    b.language,
		udid:        response.Metadata.UDID,
		startMillis: response.Metadata.StartMillis,
		model:       response.Metadata.Model,
		iosVersion:  response.Metadata.IOSVersion,
		state:       response.Session,
	}, nil
}

// UDID returns the generated uppercase device identifier for the session.
func (s *AkamaiBMPIOSSession) UDID() string {
	return s.udid
}

// StartMillis returns the session's BMP start timestamp in milliseconds.
func (s *AkamaiBMPIOSSession) StartMillis() int64 {
	return s.startMillis
}

// Model returns the selected device model identifier.
func (s *AkamaiBMPIOSSession) Model() string {
	return s.model
}

// IOSVersion returns the selected device's iOS version.
func (s *AkamaiBMPIOSSession) IOSVersion() string {
	return s.iosVersion
}

// akamaiBMPIOSSensorBuilder configures optional challenge data for one sensor
// generation.
type akamaiBMPIOSSensorBuilder struct {
	session   *AkamaiBMPIOSSession
	params    string
	dciScript string
}

// Sensor returns a builder for one sensor generation.
// Complete the builder with Generate(ctx), which makes
// one charged sensor-generation request and updates the
// session's internal state only when that request succeeds.
func (s *AkamaiBMPIOSSession) Sensor() akamaiBMPIOSSensorBuilder {
	return akamaiBMPIOSSensorBuilder{session: s}
}

// WithParams supplies the JSON response from Akamai's /_bm/get_params endpoint
// as a string
func (b akamaiBMPIOSSensorBuilder) WithParams(params string) akamaiBMPIOSSensorBuilder {
	b.params = params
	return b
}

// WithDCIScript supplies the DCI JavaScript source as a string
func (b akamaiBMPIOSSensorBuilder) WithDCIScript(script string) akamaiBMPIOSSensorBuilder {
	b.dciScript = script
	return b
}

// Generate makes one charged sensor-generation request.
func (b akamaiBMPIOSSensorBuilder) Generate(ctx context.Context) (string, ReportData, error) {
	b.session.mu.Lock()
	defer b.session.mu.Unlock()

	request := struct {
		BMPVersion string `json:"bmpVersion"`
		AppPackage string `json:"appPackage"`
		AppVersion string `json:"appVersion"`
		Language   string `json:"language,omitempty"`
		DCIScript  string `json:"dciScript,omitempty"`
		Params     string `json:"params,omitempty"`
		Session    string `json:"session"`
	}{
		BMPVersion: b.session.bmpVersion,
		AppPackage: b.session.appPackage,
		AppVersion: b.session.appVersion,
		Language:   b.session.language,
		DCIScript:  b.dciScript,
		Params:     b.params,
		Session:    b.session.state,
	}

	var response struct {
		Sensor     string     `json:"sensor"`
		Session    string     `json:"session"`
		ReportData ReportData `json:"reportData"`
	}

	if err := b.session.client.post(ctx, akamaiBMPIOSSensorEndpoint, nil, request, &response); err != nil {
		return "", "", err
	}

	b.session.state = response.Session

	return response.Sensor, response.ReportData, nil
}
