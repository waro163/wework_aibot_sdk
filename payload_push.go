package weworkaibotsdk

type PushPayload struct {
	Cmd     CMD_TYPE        `json:"cmd"`
	Headers PayloadHeaders  `json:"headers"`
	Body    PushPayloadBody `json:"body"`
}

type PushPayloadBody struct {
	MsgType      string                   `json:"msgtype"`
	Markdown     *PayloadBodyMarkdown     `json:"markdown,omitempty"`
	TemplateCard *PayloadBodyTemplateCard `json:"template_card,omitempty"`
}

type PayloadBodyMarkdown struct {
	Content string `json:"content"`
}
