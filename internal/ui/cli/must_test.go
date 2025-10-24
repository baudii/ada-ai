package cli_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/baudii/ada-ai/internal/ui/cli"
	"github.com/stretchr/testify/assert"
)

func TestMust(t *testing.T) {
	t.Parallel()
	tests := []struct {
		val any
		err error
	}{
		{"some value", nil},
		{42, nil},
		{nil, errors.New("Must failed: err")},
	}
	for _, v := range tests {
		t.Run(fmt.Sprintf("%v", v.val), func(t *testing.T) {
			if v.err == nil {
				result := cli.Must(v.val, v.err)
				assert.Equal(t, v.val, result)
			} else {
				assert.PanicsWithValue(t, v.err.Error(), func() { cli.Must(v.val, errors.New("err")) })
			}
		})
	}
}

func TestMustErr(t *testing.T) {
	t.Parallel()
	tests := []struct {
		err error
	}{
		{nil},
		{errors.New("MustErr failed: err")},
	}
	for _, v := range tests {
		t.Run(fmt.Sprintf("%v", v.err), func(t *testing.T) {
			if v.err == nil {
				cli.MustErr(nil)
			} else {
				assert.PanicsWithValue(t, v.err.Error(), func() { cli.MustErr(errors.New("err")) })
			}
		})
	}
}
