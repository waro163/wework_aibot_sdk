package weworkaibotsdk

type CallbackPayload struct {
	Cmd     string              `json:"cmd"`
	Headers PayloadHeaders      `json:"headers"`
	Body    CallbackPayloadBody `json:"body"`
}

type CallbackPayloadBody struct {
	MsgId       string           `json:"msgid"`
	AibotId     string           `json:"aibotid"`
	ChatId      string           `json:"chatid,omitempty"`
	ChatType    string           `json:"chattype"`
	From        PayloadBodyFrom  `json:"from"`
	MsgType     string           `json:"msgtype"`
	Text        PayloadBodyText  `json:"text,omitempty"`
	Image       PayloadBodyImage `json:"image,omitempty"`
	Mixed       PayloadBodyMixed `json:"mixed,omitempty"`
	Voice       PayloadBodyVoice `json:"voice,omitempty"`
	File        PayloadBodyFile  `json:"file,omitempty"`
	Quote       PayloadBodyQuote `json:"quote,omitempty"`
	ResponseUrl string           `json:"response_url,omitempty"` // 支持主动回复消息的临时 URL
	// event message
	CreateTime int64            `json:"create_time,omitempty"`
	Event      PayloadBodyEvent `json:"event,omitempty"`
	// response
	PayloadError
}

type PayloadBodyFrom struct {
	UserId string `json:"userid"`
	CorpId string `json:"corpid,omitempty"` // 事件触发者的 corpid，企业内部机器人不返回
}

type PayloadBodyText struct {
	Content string `json:"content"`
}

type PayloadBodyImage struct {
	Url     string `json:"url,omitempty"`
	AesKey  string `json:"aeskey,omitempty"`
	MediaId string `json:"mediaid,omitempty"` // 图片的 media_id
}

type PayloadBodyFile struct {
	Url     string `json:"url,omitempty"`
	AesKey  string `json:"aeskey,omitempty"`
	MediaId string `json:"mediaid,omitempty"` // 文件的 media_id
}

type PayloadBodyVoice struct {
	Content string `json:"content"`
}

type PayloadBodyMixed struct {
	MsgItem []MixedMsgItem `json:"msg_item"`
}
type MixedMsgItem struct {
	MsgType string           `json:"msgtype"`
	Text    PayloadBodyText  `json:"text"`
	Image   PayloadBodyImage `json:"image"`
}

type PayloadBodyQuote struct {
	MsgType string           `json:"msgtype"`
	Text    PayloadBodyText  `json:"text"`
	Image   PayloadBodyImage `json:"image"`
	Mixed   PayloadBodyMixed `json:"mixed"`
	Voice   PayloadBodyVoice `json:"voice"`
	File    PayloadBodyFile  `json:"file"`
}

type PayloadBodyEvent struct {
	EventType string `json:"eventtype"`
	EventKey  string `json:"event_key,omitempty"`  // 模板卡片事件：用户点击的按钮 key
	TaskId    string `json:"task_id,omitempty"`    // 模板卡片事件：任务 ID
}
