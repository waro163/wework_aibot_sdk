package weworkaibotsdk

type CallbackPayload struct {
	Cmd     string              `json:"cmd"`
	Headers PayloadHeaders      `json:"headers"`
	Body    CallbackPayloadBody `json:"body"`
}

type CallbackPayloadBody struct {
	MsgId    string             `json:"msgid"`
	AibotId  string             `json:"aibotid"`
	ChatId   string             `json:"chatid"`
	ChatType string             `json:"chattype"`
	From     PayloadBodyFrom    `json:"from"`
	MsgType  string             `json:"msgtype"`
	Text     PayloadBodyText    `json:"text"`
	Image    PayloadBodyImage   `json:"image"`
	Mixed    []PayloadBodyMixed `json:"mixed"`
	Voice    PayloadBodyVoice   `json:"voice"`
	File     PayloadBodyFile    `json:"file"`
	// event message
	CreateTime int64            `json:"create_time"`
	Event      PayloadBodyEvent `json:"event"`
	// response
	PayloadError
}

type PayloadBodyFrom struct {
	UserId string `json:"userid"`
}

type PayloadBodyText struct {
	Content string `json:"content"`
}

type PayloadBodyImage struct {
	Url    string `json:"url"`
	AesKey string `json:"aeskey"`
}

type PayloadBodyFile struct {
	Url    string `json:"url"`
	AesKey string `json:"aeskey"`
}

type PayloadBodyVoice struct {
	Content string `json:"content"`
}

type PayloadBodyMixed struct {
	MsgType string           `json:"msgtype"`
	Text    PayloadBodyText  `json:"text"`
	Image   PayloadBodyImage `json:"image"`
}

type PayloadBodyEvent struct {
	EventType string `json:"eventtype"`
}
