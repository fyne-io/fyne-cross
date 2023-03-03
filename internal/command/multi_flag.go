package command

import (
	"errors"
	"strings"
)

type multiFlags struct {
	values map[string]string
}

func (mf *multiFlags) String() string {
	s := ""
	for key, value := range mf.values {
		if s != "" {
			s += ","
		}
		s += key + "=" + value
	}

	return s
}

func (mf *multiFlags) Set(value string) error {
	splitted := strings.Split(value, "=")
	if len(splitted) != 2 {
		return errors.New("invalid metadata format, expecting key=value")
	}

	if mf.values == nil {
		mf.values = make(map[string]string)
	}

	mf.values[splitted[0]] = splitted[1]

	return nil
}
