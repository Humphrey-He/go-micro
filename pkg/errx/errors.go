package errx

const (
	CodeOK              = 0
	CodeInvalidRequest  = 40001
	CodeUnauthorized    = 40101
	CodeNotFound        = 40401
	CodeConflict        = 40901
	CodeInvalidState    = 40902
	CodeInternalError   = 50001
	CodeUpstreamUnavail = 50201
)

const (
	MsgOK              = "OK"
	MsgInvalidRequest  = "invalid request"
	MsgUnauthorized    = "missing authorization"
	MsgNotFound        = "not found"
	MsgConflict        = "conflict"
	MsgInvalidState    = "invalid state"
	MsgInternalError   = "internal error"
	MsgUpstreamUnavail = "upstream unavailable"
)
