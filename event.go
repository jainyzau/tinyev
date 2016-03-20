package socket

const (
	Event_SessionConnected = iota
	Event_SessionClosed
	Event_SessionAccepted
	Event_SessionData

	Event_AcceptorInit
	Event_AcceptorInitFailed
	Event_AcceptorInitDone
	Event_AcceptorStart
	Event_AcceptorAcceptFailed
	Event_AcceptorStop

	Event_ConnectorInit
	Event_ConnectorStart
	Event_ConnectorConnectFailed
	Event_ConnectorTryReconnecting
	Event_ConnectorConnected
	Event_ConnectorStop

	Event_PipeClosed
	Event_PipeData
)

type Event struct {
	_type int
	_data interface{}
}

type SessionData struct {
	_session *Session
	_buffer  []byte
}

type AcceptorData struct {
	_acceptor *Acceptor
	_session  *Session
}

type ConnectorData struct {
	_connector *Connector
	_error     error
}

type PipeData struct {
	_pipe   *Pipe
	_buffer []byte
}

func (this *Event) Type() int {
	return this._type
}

func (this *Event) Data() interface{} {
	return this._data
}

func (this *AcceptorData) Session() *Session {
	return this._session
}

func (this *AcceptorData) Acceptor() *Acceptor {
	return this._acceptor
}

func (this *SessionData) Session() *Session {
	return this._session
}

func (this *SessionData) Buffer() []byte {
	return this._buffer
}

func (this *ConnectorData) Connector() *Connector {
	return this._connector
}

func (this *ConnectorData) Error() string {
	return this._error.Error()
}

func (this *PipeData) Pipe() *Pipe {
	return this._pipe
}

func (this *PipeData) Buffer() []byte {
	return this._buffer
}

func NewEvent(eventType int, data interface{}) *Event {
	return &Event{
		_type: eventType,
		_data: data,
	}
}
