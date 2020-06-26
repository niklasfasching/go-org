package blorg

import "testing"

func TestBlorg(t *testing.T) {
	config, err := ReadConfig("testdata/blorg.org")
	if err != nil {
		t.Errorf("Could not read config: %s", err)
		return
	}
	if err := config.Render(); err != nil {
		t.Errorf("Could not render: %s", err)
	}
}
