package jap

import "testing"

func TestParseVlan(t *testing.T) {
	var test = "vlan 300\n name office\n!"
	var vid = 300
	vlan, err := ParseVlan(test, vid)
	if err != nil {
		t.Error(err)
	}

	if vlan.id != 300 {
		t.Errorf("Wrong VLan Id wants: `300` got: `%v`", vlan.id)
	}

	if vlan.name != "office" {
		t.Errorf("Wrong VLan name wants: `office` got: `%v`", vlan.id)
	}
}
