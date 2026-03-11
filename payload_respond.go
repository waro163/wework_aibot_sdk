package weworkaibotsdk

type RespondPayload struct {
	Cmd     CMD_TYPE           `json:"cmd"`
	Headers PayloadHeaders     `json:"headers"`
	Body    RespondPayloadBody `json:"body"`
}

type RespondPayloadBody struct {
	MsgType      string                   `json:"msgtype"`
	Text         *PayloadBodyText         `json:"text,omitempty"`
	Stream       *PayloadBodyStream       `json:"stream,omitempty"`
	ResponseType string                   `json:"response_type,omitempty"`
	TemplateCard *PayloadBodyTemplateCard `json:"template_card,omitempty"`
	UserIds      []string                 `json:"userids,omitempty"` // 更新模板卡片时使用
}

type PayloadBodyStream struct {
	Id       string          `json:"id"`
	Finish   bool            `json:"finish,omitempty"`
	Content  string          `json:"content,omitempty"`
	MsgItem  []StreamMsgItem `json:"msg_item,omitempty"` // 图文混排消息列表，目前仅当 finish=true 时支持设置
	Feedback *StreamFeedback `json:"feedback,omitempty"` // 反馈信息
}

type StreamMsgItem struct {
	MsgType string           `json:"msgtype"`
	Image   StreamImageItem  `json:"image,omitempty"`
}

type StreamImageItem struct {
	Base64 string `json:"base64"` // Base64 编码的图片数据
	Md5    string `json:"md5"`    // 图片内容的 MD5 值
}

type StreamFeedback struct {
	Id string `json:"id"` // 反馈 ID
}

type PayloadBodyTemplateCard struct {
	CardType              string                         `json:"card_type"`
	Source                *TemplateCardSource            `json:"source,omitempty"`
	ActionMenu            *TemplateCardActionMenu        `json:"action_menu,omitempty"`
	MainTitle             *TemplateCardMainTitle         `json:"main_title,omitempty"`
	EmphasisContent       *TemplateCardEmphasisContent   `json:"emphasis_content,omitempty"`
	QuoteArea             *TemplateCardQuoteArea         `json:"quote_area,omitempty"`
	SubTitleText          string                         `json:"sub_title_text,omitempty"`
	HorizontalContentList []TemplateCardHorizontalContent `json:"horizontal_content_list,omitempty"`
	JumpList              []TemplateCardJumpAction       `json:"jump_list,omitempty"`
	CardAction            *TemplateCardAction            `json:"card_action,omitempty"`
	CardImage             *TemplateCardImage             `json:"card_image,omitempty"`
	ImageTextArea         *TemplateCardImageTextArea     `json:"image_text_area,omitempty"`
	VerticalContentList   []TemplateCardVerticalContent  `json:"vertical_content_list,omitempty"`
	ButtonSelection       *TemplateCardSelectionItem     `json:"button_selection,omitempty"`
	ButtonList            []TemplateCardButton           `json:"button_list,omitempty"`
	Checkbox              *TemplateCardCheckbox          `json:"checkbox,omitempty"`
	SelectList            []TemplateCardSelectionItem    `json:"select_list,omitempty"`
	SubmitButton          *TemplateCardSubmitButton      `json:"submit_button,omitempty"`
	TaskId                string                         `json:"task_id,omitempty"`
	Feedback              *TemplateFeedback              `json:"feedback,omitempty"`
}

// 卡片来源样式信息
type TemplateCardSource struct {
	IconUrl   string `json:"icon_url,omitempty"`
	Desc      string `json:"desc,omitempty"`
	DescColor int    `json:"desc_color,omitempty"` // 0(默认)灰色，1 黑色，2 红色，3 绿色
}

// 卡片右上角更多操作按钮
type TemplateCardActionMenu struct {
	Desc       string                     `json:"desc"`
	ActionList []TemplateCardActionItem   `json:"action_list"`
}

type TemplateCardActionItem struct {
	Text string `json:"text"`
	Key  string `json:"key"`
}

// 模板卡片主标题
type TemplateCardMainTitle struct {
	Title string `json:"title,omitempty"`
	Desc  string `json:"desc,omitempty"`
}

// 关键数据样式
type TemplateCardEmphasisContent struct {
	Title string `json:"title,omitempty"`
	Desc  string `json:"desc,omitempty"`
}

// 引用文献样式
type TemplateCardQuoteArea struct {
	Type      int    `json:"type,omitempty"`      // 0 或不填代表没有点击事件，1 代表跳转 url，2 代表跳转小程序
	Url       string `json:"url,omitempty"`
	Appid     string `json:"appid,omitempty"`
	Pagepath  string `json:"pagepath,omitempty"`
	Title     string `json:"title,omitempty"`
	QuoteText string `json:"quote_text,omitempty"`
}

// 二级标题+文本列表
type TemplateCardHorizontalContent struct {
	Type    int    `json:"type,omitempty"` // 0 或不填代表普通文本，1 代表跳转 url，3 代表点击跳转成员详情
	Keyname string `json:"keyname"`
	Value   string `json:"value,omitempty"`
	Url     string `json:"url,omitempty"`
	UserId  string `json:"userid,omitempty"`
}

// 跳转指引样式
type TemplateCardJumpAction struct {
	Type     int    `json:"type,omitempty"` // 0 或不填代表不是链接，1 代表跳转 url，2 代表跳转小程序，3 代表触发消息智能回复
	Title    string `json:"title"`
	Url      string `json:"url,omitempty"`
	Appid    string `json:"appid,omitempty"`
	Pagepath string `json:"pagepath,omitempty"`
	Question string `json:"question,omitempty"`
}

// 整体卡片的点击跳转事件
type TemplateCardAction struct {
	Type     int    `json:"type"` // 0 或不填代表不是链接，1 代表跳转 url，2 代表打开小程序
	Url      string `json:"url,omitempty"`
	Appid    string `json:"appid,omitempty"`
	Pagepath string `json:"pagepath,omitempty"`
}

// 卡片二级垂直内容
type TemplateCardVerticalContent struct {
	Title string `json:"title"`
	Desc  string `json:"desc,omitempty"`
}

// 图片样式
type TemplateCardImage struct {
	Url         string  `json:"url"`
	AspectRatio float64 `json:"aspect_ratio,omitempty"`
}

// 左图右文样式
type TemplateCardImageTextArea struct {
	Type     int    `json:"type,omitempty"` // 0 或不填代表没有点击事件，1 代表跳转 url，2 代表跳转小程序
	Url      string `json:"url,omitempty"`
	Appid    string `json:"appid,omitempty"`
	Pagepath string `json:"pagepath,omitempty"`
	Title    string `json:"title,omitempty"`
	Desc     string `json:"desc,omitempty"`
	ImageUrl string `json:"image_url"`
}

// 下拉式选择器
type TemplateCardSelectionItem struct {
	QuestionKey string                      `json:"question_key"`
	Title       string                      `json:"title,omitempty"`
	Disable     bool                        `json:"disable,omitempty"`
	SelectedId  string                      `json:"selected_id,omitempty"`
	OptionList  []TemplateCardSelectionOption `json:"option_list"`
}

type TemplateCardSelectionOption struct {
	Id   string `json:"id"`
	Text string `json:"text"`
}

// 模板卡片按钮
type TemplateCardButton struct {
	Text  string `json:"text"`
	Style int    `json:"style,omitempty"` // 1~4，不填或错填默认 1
	Key   string `json:"key"`
}

// 选择题样式（投票选择）
type TemplateCardCheckbox struct {
	QuestionKey string                       `json:"question_key"`
	Disable     bool                         `json:"disable,omitempty"`
	Mode        int                          `json:"mode,omitempty"` // 单选：0，多选：1，不填默认 0
	OptionList  []TemplateCardCheckboxOption `json:"option_list"`
}

type TemplateCardCheckboxOption struct {
	Id        string `json:"id"`
	Text      string `json:"text"`
	IsChecked bool   `json:"is_checked,omitempty"`
}

// 提交按钮样式
type TemplateCardSubmitButton struct {
	Text string `json:"text"`
	Key  string `json:"key"`
}

// 反馈信息
type TemplateFeedback struct {
	Id string `json:"id"`
}
