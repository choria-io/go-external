package agent

// StatusCode is a reply status as defined by MCollective SimpleRPC - integers 0 to 5
type StatusCode uint8

const (
	// OK is the reply status when all worked
	OK = StatusCode(iota)

	// Aborted is status for when the action could not run, most failures in an action should set this
	Aborted

	// UnknownAction is the status for unknown actions requested
	UnknownAction

	// MissingData is the status for missing input data
	MissingData

	// InvalidData is the status for invalid input data
	InvalidData

	// UnknownError is the status general failures in agents should set when things go bad
	UnknownError
)

// Reply is the reply data as stipulated by MCollective RPC system.  The Data
// has to be something that can be turned into JSON using the normal Marshal system
type Reply struct {
	Statuscode StatusCode  `json:"statuscode"`
	Statusmsg  string      `json:"statusmsg"`
	Data       interface{} `json:"data"`
}

// Abort sets the status code and message of the RPC reply
func (r *Reply) Abort(c StatusCode, msg string) {
	r.Statuscode = c
	r.Statusmsg = msg
}
