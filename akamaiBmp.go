package veris

import "encoding/json"

type AkamaiBMPParams struct {
	ServerSideSignal string `json:"serversidesignal"`
}

func ParseAkamaiBMPParams(paramsBytes []byte) (AkamaiBMPParams, error) {
	var params AkamaiBMPParams
	if err := json.Unmarshal(paramsBytes, &params); err != nil {
		return AkamaiBMPParams{}, err
	}

	return params, nil
}
