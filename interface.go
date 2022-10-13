package jap

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type CiscoInterface struct {
	Identifier         string
	SubInterface       int
	Description        string `cmd:"description ([[:print:]]+)"`
	Shutdown           bool   `cmd:"shutdown"`
	Trunk              bool   `cmd:"switchport mode trunk"`
	AccessPort         bool   `cmd:"switchport mode access"`
	NativeVlan         int    `cmd:"switchport trunk native vlan ([0-9]+)"`
	Ips                []Ip
	Vrf                string `cmd:"ip vrf forwarding ([[:print:]]+)"`
	DhcpSnoopingThrust bool   `cmd:"ip dhcp snooping trust"`
	AccessVlan         int    `cmd:"switchport access vlan ([0-9]+)"`
	VoiceVlan          int    `cmd:"switchport voice vlan ([0-9]+)"`
	IPHelperAddresses  []string
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

	//IP Helper Addresses
	re = regexp.MustCompile(`ip helper-address (\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})`)
	ips = re.FindAllStringSubmatch(part, -1)
	for _, intIp := range ips {
		inter.IPHelperAddresses = append(inter.IPHelperAddresses, intIp[1])
	}

	//
	// Get all the rest stuff
	//
	t := reflect.TypeOf(inter)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("cmd")
		if tag != "" {
			re = regexp.MustCompile(tag)
			// @todo check if no is with the command!
			data := re.FindStringSubmatch(part)
			if len(data) > 1 && field.Type.Kind() == reflect.String {
				reflect.ValueOf(&inter).Elem().Field(i).SetString(data[1])
			} else if len(data) > 1 && field.Type.Kind() == reflect.Int {
				value, err := strconv.ParseInt(data[1], 10, 64)
				if err != nil {
					return CiscoInterface{}, err
				}
				reflect.ValueOf(&inter).Elem().Field(i).SetInt(value)
			} else if len(data) == 1 && field.Type.Kind() == reflect.Bool {
				reflect.ValueOf(&inter).Elem().Field(i).SetBool(true)
			}
		}
	}

	//Check if Routed Port, Trunk or Access
	if !inter.AccessPort && !inter.Trunk {
		if !strings.Contains(part, "ip address") && !strings.Contains(part, "no switchport") {
			inter.AccessPort = true
		}
	}

	return inter, nil
}
