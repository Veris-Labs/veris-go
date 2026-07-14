package veris

import "context"

type ReportData string

func (client *Client) Feedback(ctx context.Context, valid bool, reportData ReportData) error {
	reqBody := struct {
		Valid      bool       `json:"valid"`
		ReportData ReportData `json:"reportData"`
	}{
		Valid:      valid,
		ReportData: reportData,
	}

	return client.post(ctx, feedbackEndpoint, nil, reqBody, nil)
}
