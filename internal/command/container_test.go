package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_concatNonEmptyNamedArgs(t *testing.T) {
	assert.Equal(t, []string{"-foo", "42", "-baz", "64"}, concatNonEmptyNamedArgs(
		"-foo", "42",
		"-bar", "",
		"-baz", "64",
	), "filter empty argument/value pair")
	assert.Equal(t, []string{"-foo", "42"}, concatNonEmptyNamedArgs(
		"-foo", "42",
		"-bar", "",
		"-baz",
	), "trailing/incomplete/boolean argument")
}
