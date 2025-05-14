package errs

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"runtime"
	"strconv"
	"testing"
)

func TestErr(t *testing.T) {
	filePath1, err1 := caller1(t)

	assert.Error(t, err1)

	filePath2, err2 := caller2(t, err1)

	assert.Error(t, err2)

	assert.Contains(t, err1.Error(), filePath1)
	assert.NotContains(t, err1.Error(), filePath2)

	assert.Contains(t, err2.Error(), filePath1)
	assert.Contains(t, err2.Error(), filePath2)
}

func caller1(t *testing.T) (string, error) {
	_, file, line, ok := runtime.Caller(0)

	assert.True(t, ok)

	expectedFileLine := line + 6

	return file + ":" + strconv.Itoa(expectedFileLine), Err(errors.New("error message 1"))
}

func caller2(t *testing.T, err error) (string, error) {
	_, file, line, ok := runtime.Caller(0)

	assert.True(t, ok)

	expectedFileLine := line + 6

	return file + ":" + strconv.Itoa(expectedFileLine), Err(errors.New(fmt.Sprintf("%s: %s", err, "error message 2")))
}
