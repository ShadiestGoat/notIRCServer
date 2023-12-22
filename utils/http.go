package utils

type HTTPStatus interface {
	GetStatus() int
}

type HTTPErr struct {
	Msg    string
	Status int
}

func (r HTTPErr) Error() string {
	return r.Msg
}

func (r HTTPErr) GetStatus() int {
	return r.Status
}

var (
	ErrBadPerms = HTTPErr{
		Msg:    "Bad Perms <3",
		Status: 403,
	}
	ErrNotAuthorized = HTTPErr{
		Msg:    "Not Authorized <3",
		Status: 401,
	}
)
