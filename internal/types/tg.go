package types

type ITG interface {
	Start() func(error)
}

/* const (
	COMMAND_INFO        = "info"
	COMMAND_LOGLEVEL    = "loglevel"
	COMMAND_TESTSTREAM  = "teststream"
	COMMAND_AUTOFORWARD = "autoforward"
	COMMAND_AUTODEL     = "autodel"
	COMMAND_POST        = "post"
	COMMAND_GETADMINS   = "getadmins"
	COMMAND_SENDMSG     = "sendmsg"
	COMMAND_DELALL      = "delall"
	COMMAND_HELP        = "help"

	AFORW_DEBUG = "DEBUG"
	AFORW_OFF   = "OFF"
	AFORW_ON    = "ON"

	ADEL_OFF = "OFF"
	ADEL_ON  = "ON"
)
*/
