package weworkaibotsdk

type RespondPayload struct {
	Cmd     CMD_TYPE           `json:"cmd"`
	Headers PayloadHeaders     `json:"headers"`
	Body    RespondPayloadBody `json:"body"`
}

type RespondPayloadBody struct {
	MsgType      string                  `json:"msgtype"`
	Text         PayloadBodyText         `json:"text"`
	Stream       PayloadBodyStream       `json:"stream"`
	ResponseType string                  `json:"response_type"`
	TemplateCard PayloadBodyTemplateCard `json:"template_card"`
}

type PayloadBodyStream struct {
	Id      string `json:"id"`
	Finish  bool   `json:"finish"`
	Content string `json:"content"`
}

type PayloadBodyTemplateCard struct {
	CardType   string                `json:"card_type"`
	MainTitle  TemplateCardMainTitle `json:"main_title"`
	ButtonList []TemplateCardButton  `json:"button_list"`
	TaskId     string                `json:"task_id"`
	FeedBack   TemplateFeedBack      `json:"feedback"`
}

type TemplateCardMainTitle struct {
	Title string `json:"title"`
	Desc  string `json:"desc"`
}

type TemplateCardButton struct {
	Text  string `json:"text"`
	Stype int    `json:"stype"`
	Key   string `json:"key"`
}

type TemplateFeedBack struct {
	Id string `json:"id"`
}
