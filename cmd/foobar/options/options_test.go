package options

import (
	"testing"
)

func Test_parse(t *testing.T) {
	c, err := Load("../../../conf/dev/appconfig.yaml")
	if err != nil {
		t.Errorf("%+v", err)
		return
	}

	t.Logf("%+v", c)
}
