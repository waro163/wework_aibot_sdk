package weworkaibotsdk

type Config struct {
	BotId                     string `json:"bot_id"`
	Secret                    string `json:"secret"`
	WsConnectionTimeout       int    `json:"connection_timeout"`        // seconds
	WsConnectionNeedRetry     bool   `json:"connection_need_retry"`
	WsConnectionRetryInterval int    `json:"connection_retry_interval"` // seconds
	WsConnectionMaxRetryTimes int    `json:"connection_max_retry_times"`
	WsHandshakeTimeout        int    `json:"handshake_timeout"`         // seconds
	HeartbeatInterval         int    `json:"heartbeat_interval"`        // seconds, default 30
	AutoReconnect             bool   `json:"auto_reconnect"`            // enable auto-reconnect mode
}
