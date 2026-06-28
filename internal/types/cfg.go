package types

type ICfg interface {
	GetEnvVal(string) string
	SetEnvVal(string, string)
	GetJsonUsers() []User
	GetJsonAdmin() User
	Parse() ICfg
}

const (
	TS_APP_NAME  = "TS_APP_NAME"
	TS_APP_IP    = "TS_APP_IP"
	TS_LOG_LEVEL = "TS_LOG_LEVEL"
	TG_BOT_PROXY = "TG_BOT_PROXY"
	TG_BOT_TOKEN = "TG_BOT_TOKEN"
)

type User struct {
	Nickname string
	TgUserID string
}

type JsonVals struct {
	Admin User
	Users []User
}
