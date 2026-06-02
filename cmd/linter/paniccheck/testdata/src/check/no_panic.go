package check

import "errors"

func ReturnsError() error {
	return errors.New("fail")
}
