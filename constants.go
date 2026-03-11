package weworkaibotsdk

type CMD_TYPE string

const (
	WEBSOCKET_URL = "wss://openws.work.weixin.qq.com"

	ERROR_CODE_SUCCESS = 0
	ERROR_MSG_SUCCESS  = "ok"

	CMD_AIBOT_MSG_CALLBACK                 = "aibot_msg_callback"
	CMD_AIBOT_EVENT_CALLBACK               = "aibot_event_callback"
	CMD_AIBOT_RESPOND_WELCOME_MSG CMD_TYPE = "aibot_respond_welcome_msg"
	CMD_AIBOT_RESPOND_MSG         CMD_TYPE = "aibot_respond_msg"
	CMD_AIBOT_RESPOND_UPDATE_MSG  CMD_TYPE = "aibot_respond_update_msg"
	CMD_AIBOT_SEND_MSG            CMD_TYPE = "aibot_send_msg"
	CMD_AIBOT_SUBSCRIBE           CMD_TYPE = "aibot_subscribe"
	CMD_PING                      CMD_TYPE = "ping"

	MSG_TYPE_TEXT   = "text"
	MSG_TYPE_IMAGE  = "image"
	MSG_TYPE_VOICE  = "voice"
	MSG_TYPE_MIXED  = "mixed"
	MSG_TYPE_FILE   = "file"
	MSG_TYPE_EVENT  = "event"
	MSG_TYPE_STREAM = "stream"

	EVENT_TYPE_ENTER_CHAT    = "enter_chat"
	EVENT_TYPE_TEMPLATE_CARD = "template_card_event"
	EVENT_TYPE_FEEDBACK      = "feedback_event"
	EVENT_TYPE_DISCONNECTED  = "disconnected_event"
)
