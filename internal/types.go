package internal

// Request struct
type Request struct {
	Tag      string `json:"tag,omitempty"` // unique client identifier
	Action   string `json:"action"`        // action to perform
	Data     string `json:"data"`          // data to send back
	Generate struct {
		Model   string `json:"model"`
		Prompt  string `json:"prompt"`
		Context []int  `json:"context,omitempty"`
	} `json:"generate"`
}
