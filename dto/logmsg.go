package dto

import "encoding/json"

type LogMsg struct {
	Operation string `json:"OPERATION,omitempty"`
	From      string `json:"FROM,omitempty"`
	To        string `json:"TO,omitempty"`
	// RC        string `json:"RC,omitempty"`
	// CRN    string `json:"CRN,omitempty"`
	RawMsg string `json:"RAW_MSG,omitempty"`
}

func BuildLogRequest(operation string, dtoReq interface{}) LogMsg {
	jsonMsg, err := json.Marshal(dtoReq)
	rawMsg := string(jsonMsg)
	if err != nil {
		rawMsg = err.Error()
	}

	return LogMsg{
		Operation: operation,
		From:      "BNI", To: "API",
		// RC: rc, CRN: crn,
		RawMsg: rawMsg,
	}
}
func BuildLogResponse(operation string, dtoResp interface{}) LogMsg {
	jsonMsg, err := json.Marshal(dtoResp)
	rawMsg := string(jsonMsg)
	if err != nil {
		rawMsg = err.Error()
	}

	return LogMsg{
		Operation: operation,
		From:      "API", To: "BNI",
		// RC: rc, CRN: crn,
		RawMsg: rawMsg,
	}
}
