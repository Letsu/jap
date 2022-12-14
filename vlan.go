package jap

import "regexp"

type Vlan struct {
	Id   int
	Name string
}

func ParseVlan(part string, vlanId int) (Vlan, error) {
	var vlan Vlan
	vlan.Id = vlanId
	re := regexp.MustCompile(`name ([[:print:]]+)`)
	fullName := re.FindStringSubmatch(part)
	if len(fullName) > 0 {
		vlan.Name = fullName[1]
	}

	return vlan, nil
}
