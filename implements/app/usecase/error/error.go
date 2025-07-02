package error

type InvalidParameter struct {
	message string
	err     error
}

func (e InvalidParameter) Error() string {
	return e.message
}

func NewInvalidParameter(message string, err error) error {
	return InvalidParameter{
		message: message,
		err:     err,
	}
}

type NotFound struct {
	message string
	err     error
}

func (e NotFound) Error() string {
	return e.message
}

func NewNotFound(message string, err error) error {
	return NotFound{
		message: message,
		err:     err,
	}
}

type AlreadyExists struct {
	message string
	err     error
}

func (e AlreadyExists) Error() string {
	return e.message
}

func NewAlreadyExists(message string, err error) error {
	return AlreadyExists{
		message: message,
		err:     err,
	}
}

type PermissionDenied struct {
	message string
	err     error
}

func (e PermissionDenied) Error() string {
	return e.message
}

func NewNewPermissionDenied(message string, err error) error {
	return PermissionDenied{
		message: message,
		err:     err,
	}
}
