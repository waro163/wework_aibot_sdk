package weworkaibotsdk

type PayloadHeaders struct {
	ReqId string `json:"req_id"`
}

type PayloadError struct {
	ErrCode *int   `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}
