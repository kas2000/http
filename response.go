package http

type Response interface {
	Headers() map[string]string
	SetHeader(key, value string)
	GetHeader(key string) string

	Response() interface{}
	StatusCode() int
}

type response struct {
	Status int
	Body interface{}
	HeaderData map[string]string
}

func NewResponse(status int, body interface{}, headerData map[string]string) Response {
	return &response{Status: status, Body: body, HeaderData: headerData}
}

func (e *response) Response() interface{} {
	return e.Body
}

func (e *response) StatusCode() int {
	return e.Status
}

func (e *response) Headers() map[string]string {
	return e.HeaderData
}

func (e *response) SetHeader(key, value string) {
	if e.HeaderData == nil {
		e.HeaderData = map[string]string{}
	}
	e.HeaderData[key] = value
}

func (e *response) GetHeader(key string) string {
	return e.HeaderData[key]
}