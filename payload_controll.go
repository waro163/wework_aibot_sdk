package weworkaibotsdk

type AuthPayload struct {
	Cmd     CMD_TYPE        `json:"cmd"`
	Headers PayloadHeaders  `json:"headers"`
	Body    AuthPayloadBody `json:"body"`
}

type AuthPayloadBody struct {
	BotId  string `json:"bot_id"`
	Secret string `json:"secret"`
}

type PingPayload struct {
	Cmd     CMD_TYPE       `json:"cmd"`
	Headers PayloadHeaders `json:"headers"`
}
