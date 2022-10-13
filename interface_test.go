package jap

import (
	"io/ioutil"
	"testing"
)

type TestInterface struct {
	FileName           string
	Identifier         string
	Description        string
	SubInterface       int
	Shutdown           bool
	Trunk              bool
	NativeVlan         int
	TestIps            []TestIp
	Vrf                string
	DhcpSnoopingThrust bool
	AccessPort         bool
	AccessVlan         int
	VoiceVlan          int
	IPHelperAddresses  []string
}

type TestIp struct {
	IpAdd     string
	Subnet    string
	Secondary bool
}

func TestParseInterface(t *testing.T) {
	var testInters []TestInterface
	// Layer 3 interface
	ip1 := TestIp{
		IpAdd:     "10.0.2.1",
		Subnet:    "255.255.255.0",
		Secondary: false,
	}
	ip2 := TestIp{
		IpAdd:     "192.1.0.2",
		Subnet:    "255.255.255.0",
		Secondary: true,
	}
	layer3 := TestInterface{
		FileName:           "layer3.txt",
		Identifier:         "BVI2",
		Description:        "",
		Shutdown:           false,
		Trunk:              false,
		TestIps:            nil,
		Vrf:                "red-custom",
		DhcpSnoopingThrust: false,
		AccessPort:         false,
		IPHelperAddresses:  []string{"10.0.0.100", "10.0.0.101"},
	}
	layer3.TestIps = append(layer3.TestIps, ip1)
	layer3.TestIps = append(layer3.TestIps, ip2)
	testInters = append(testInters, layer3)

	// Layer 2 trunk interface
	layer2trunk := TestInterface{
		FileName:           "layer2trunk.txt",
		Identifier:         "TenGigabitEthernet1/0/2",
		Description:        "gi1/0/2@cr-01",
		SubInterface:       0,
		Shutdown:           false,
		Trunk:              true,
		NativeVlan:         188,
		TestIps:            nil,
		Vrf:                "",
		DhcpSnoopingThrust: true,
		AccessPort:         false,
		AccessVlan:         0,
		VoiceVlan:          0,
	}
	testInters = append(testInters, layer2trunk)

	// Layer 2 access interface
	layer2access := TestInterface{
		FileName:           "layer2access.txt",
		Identifier:         "FastEthernet0/2",
		Description:        "Access to VLAN 200",
		SubInterface:       0,
		Shutdown:           true,
		Trunk:              false,
		NativeVlan:         0,
		TestIps:            nil,
		Vrf:                "",
		DhcpSnoopingThrust: false,
		AccessPort:         true,
		AccessVlan:         200,
		VoiceVlan:          300,
	}
	testInters = append(testInters, layer2access)

	for _, testInter := range testInters {
		// Open File
		content, err := ioutil.ReadFile("testFiles/interface/" + testInter.FileName)
		if err != nil {
			t.Error(err)
		}

		inter, err := ParseInterface(string(content))
		if err != nil {
			t.Error(err)
		}

		if len(testInter.TestIps) != len(inter.Ips) {
			t.Errorf("Wrong number of interfaces returned want: `%d` got: `%d`", len(testInter.TestIps), len(inter.Ips))
		}

		if testInter.Identifier != inter.Identifier {
			t.Errorf("Got wrong interface identifier want: `%s`, got: `%s`", testInter.Identifier, inter.Identifier)
		}

		if testInter.SubInterface != inter.SubInterface {
			t.Errorf("Got wrong subinterface want: `%d`, got: `%d`", testInter.SubInterface, inter.SubInterface)
		}

		if testInter.Description != inter.Description {
			t.Errorf("Got wrong descripton want: %s, got: %s", testInter.Description, inter.Description)
		}

		if testInter.Shutdown != inter.Shutdown {
			t.Errorf("Got wrong shutdown state want: `%t`, got: `%t`", testInter.Shutdown, inter.Shutdown)
		}

		if testInter.Trunk != inter.Trunk {
			t.Errorf("Got wrong trunk state want: `%t`, got: `%t`", testInter.Trunk, inter.Trunk)
		}

		if testInter.NativeVlan != inter.NativeVlan {
			t.Errorf("Got wrong native vlan want: `%d`, got: `%d`", testInter.NativeVlan, inter.NativeVlan)
		}

		if testInter.Vrf != inter.Vrf {
			t.Errorf("Got wrong vrf want: `%s`, got: `%s`", testInter.Vrf, inter.Vrf)
		}

		if testInter.DhcpSnoopingThrust != inter.DhcpSnoopingThrust {
			t.Errorf("Got wrong dhcp snooping thrust mode want: `%t`, got: `%t`", testInter.DhcpSnoopingThrust, inter.DhcpSnoopingThrust)
		}

		if testInter.AccessPort != inter.AccessPort {
			t.Errorf("Got wrong access port want: `%t`, got: `%t` on %s, %s", testInter.AccessPort, inter.AccessPort, testInter.FileName, inter.Identifier)
		}

		if testInter.AccessVlan != inter.AccessVlan {
			t.Errorf("Got wrong access vlan want: `%d`, got: `%d`", testInter.AccessVlan, inter.AccessVlan)
		}

		if testInter.VoiceVlan != inter.VoiceVlan {
			t.Errorf("Got wrong voice vlan want: `%d`, got: `%d`", testInter.VoiceVlan, inter.VoiceVlan)
		}

		for i, helperAddress := range inter.IPHelperAddresses {
			found := false
			for _, ip := range testInter.IPHelperAddresses {
				if ip == helperAddress {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Did not find ip helper: `%s` in interface: `%s`", testInter.IPHelperAddresses[i], testInter.Identifier)
			}
		}

		for i, ip := range testInter.TestIps {
			if ip.IpAdd != inter.Ips[i].Ip {
				t.Errorf("Got wrong ip address want: `%s` got `%s`", ip.IpAdd, inter.Ips[i].Ip)
			}

			if ip.Subnet != inter.Ips[i].Subnet {
				t.Errorf("Got wrong subnet address want: `%s` got `%s`", ip.Subnet, inter.Ips[i].Subnet)
			}

			if ip.Secondary != inter.Ips[i].Secondary {
				t.Errorf("Got wrong secondary address want: `%t` got `%t`", ip.Secondary, inter.Ips[i].Secondary)
			}
		}
	}
}
