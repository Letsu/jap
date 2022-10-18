package jap

import "testing"

func TestParseVlan(t *testing.T) {
	var test = "vlan 300\n name office\n!"
	var vid = 300
	vlan, err := ParseVlan(test, vid)
	if err != nil {
		t.Error(err)
	}

	if vlan.Id != 300 {
		t.Errorf("Wrong VLan Id wants: `300` got: `%v`", vlan.Id)
	}

	if vlan.Name != "office" {
		t.Errorf("Wrong VLan name wants: `office` got: `%v`", vlan.Id)
	}
}
