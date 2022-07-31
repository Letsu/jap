package jap

import (
	"regexp"
	"strconv"
	"strings"
)

type CiscoInterface struct {
	Identifier         string
	SubInterface       int
	Description        string
	Shutdown           bool
	Trunk              bool
	AccessPort         bool
	NativeVlan         int
	Ips                []Ip
	Vrf                string
	DhcpSnoopingThrust bool
	AccessVlan         int
	VoiceVlan          int
	IPHHelperAddresses []string
}

type Ip struct {
	Ip        string
	Subnet    string
	Secondary bool
	VRF       string
}

func ParseInterface(part string) (CiscoInterface, error) {
	var inter CiscoInterface
	// Get identifier
	re := regexp.MustCompile(`interface ([\w\/\.\-\:]+)`)
	identifier := re.FindStringSubmatch(part)
	identifier = strings.Split(identifier[1], ".")
	inter.Identifier = identifier[0]
	if len(identifier) > 1 {
		inter.SubInterface, _ = strconv.Atoi(identifier[1])
	}

	// Get description
	re = regexp.MustCompile(`description ([[:print:]]+)`)
	descriptionPart := re.FindStringSubmatch(part)
	if len(descriptionPart) > 0 {
		inter.Description = descriptionPart[1]
	}

	// Get port shutdown
	re = regexp.MustCompile(`shutdown`)
	inter.Shutdown = false
	if re.MatchString(part) {
		inter.Shutdown = true
	}

	// Get ipv4 addresses
	ipRe := regexp.MustCompile(`(?m)ip address (\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}) (\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})( secondary)?( vrf ([\w\-]+))?`)
	ips := ipRe.FindAllStringSubmatch(part, -1)
	for _, intIp := range ips {
		ipAdd := Ip{
			Ip:        intIp[1],
			Subnet:    intIp[2],
			Secondary: false,
		}

		if strings.Contains(intIp[3], "secondary") {
			ipAdd.Secondary = true
		}

		if len(intIp) > 3 {
			ipAdd.VRF = intIp[5]
		}

		inter.Ips = append(inter.Ips, ipAdd)
	}

	// VRF Forwarding
	re = regexp.MustCompile(`ip vrf forwarding ([[:print:]]+)`)
	vrf := re.FindStringSubmatch(part)
	if len(vrf) > 0 {
		inter.Vrf = vrf[1]
	}

	//IP Helper Addresses
	re = regexp.MustCompile(`ip helper-address (\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})`)
	ips = re.FindAllStringSubmatch(part, -1)
	for _, intIp := range ips {
		inter.IPHHelperAddresses = append(inter.IPHHelperAddresses, intIp[1])
	}

	//
	// Get Trunk stuff
	//
	if strings.Contains(part, "switchport mode trunk") && !strings.Contains(part, "switchport mode access") {
		inter.Trunk = true
	}

	//Native Vlan
	re = regexp.MustCompile(`switchport trunk native vlan ([0-9]+)`)
	nativeVlan := re.FindStringSubmatch(part)
	if len(nativeVlan) > 0 {
		inter.NativeVlan, _ = strconv.Atoi(nativeVlan[1])
	}

	// IP dhcp snopping thrust
	inter.DhcpSnoopingThrust = false
	if strings.Contains(part, "ip dhcp snooping trust") {
		inter.DhcpSnoopingThrust = true
	}

	//
	// Get Access Stuff
	//

	// Access state
	if strings.Contains(part, "switchport mode access") && !strings.Contains(part, "switchport mode trunk") {
		inter.AccessPort = true
	}

	//Access vlan
	re = regexp.MustCompile(`switchport access vlan ([0-9]+)`)
	accessVlan := re.FindStringSubmatch(part)
	if len(accessVlan) > 0 {
		inter.AccessVlan, _ = strconv.Atoi(accessVlan[1])
	}

	// Voice Vlan
	re = regexp.MustCompile(`switchport voice vlan ([0-9]+)`)
	voiceVlan := re.FindStringSubmatch(part)
	if len(voiceVlan) > 0 {
		inter.VoiceVlan, _ = strconv.Atoi(voiceVlan[1])
	}

	return inter, nil
}
