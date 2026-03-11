package weworkaibotsdk

type PushPayload struct {
	Cmd     CMD_TYPE        `json:"cmd"`
	Headers PayloadHeaders  `json:"headers"`
	Body    PushPayloadBody `json:"body"`
}

type PushPayloadBody struct {
	Markdown PayloadBodyMarkdown `json:"markdown"`
}

type PayloadBodyMarkdown struct {
	Content string `json:"content"`
}
