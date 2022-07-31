package jap

import "regexp"

type Vlan struct {
	id   int
	name string
}

func ParseVlan(part string, vlanId int) (Vlan, error) {
	var vlan Vlan
	vlan.id = vlanId
	re := regexp.MustCompile(`name ([[:print:]]+)`)
	fullName := re.FindStringSubmatch(part)
	if len(fullName) > 0 {
		vlan.name = fullName[1]
	}

	return vlan, nil
}
