package jap

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// CiscoInterface contains all regex for searching as field Tag. The values if the field defines how the data is added.
// bool: if regex is found, sets variable to true
// int, string, float64: just adds the first found value of the first capture group to the data.
// []string: searches for all occurrences of the string in the config and adds the first capture group of the snippet to the array.
// []int: handles cisco number list and adds all capture groups of the regex to the data.
// struct: its just called the same function for parsing and are for complexer part of config which contains multiple different values in one line.
//
// The cmd tag is for config generation. It uses sprintf for getting the config back together so format values needs to be used.
type CiscoInterface struct {
	Identifier            string
	SubInterface          int
	AccessVlan            int      `reg:"switchport access vlan ([0-9]+)" cmd:"switchport access vlan %d"`
	Access                bool     `reg:"switchport mode access" cmd:"switchport mode access"`
	VoiceVlan             int      `reg:"switchport voice vlan ([0-9]+)" cmd:"switchport voice vlan %d"`
	PortSecurityMaximum   int      `reg:"switchport port-security maximum ([0-9]+)" cmd:"switchport port-security maximum %d"`
	PortSecurityViolation string   `reg:"switchport port-security violation (protect|restrict|shutdown)" cmd:"switchport port-security violation %s"`
	PortSecurityAgingTime int      `reg:"switchport port-security aging time ([0-9]+)" cmd:"switchport port-security aging time %d"`
	PortSecurityAgingType string   `reg:"switchport port-security aging type (absolute|inactivity)" cmd:"switchport port-security type  %s"`
	PortSecurity          bool     `reg:"switchport port-security" cmd:"switchport port-security"`
	Description           string   `reg:"description ([[:print:]]+)" cmd:"description %s"`
	NativeVlan            int      `reg:"switchport trunk native vlan ([0-9]+)" cmd:"switchport trunk native vlan %d"`
	TrunkAllowedVlan      []int    `reg:"switchport trunk allowed vlan( add)? ([\\d,-]+)"`
	Trunk                 bool     `reg:"switchport mode trunk" cmd:"switchport mode trunk"`
	Shutdown              bool     `reg:"shutdown" cmd:"shutdown"`
	SCBroadcastLevel      float64  `reg:"storm-control broadcast level ([0-9\\.]+)" cmd:"storm-control broadcast level %.2f"`
	STPPortFast           string   `reg:"spanning-tree portfast (disable|edge|network)" cmd:"spanning-tree portfast %s"`
	STPBpduGuard          string   `reg:"spanning-tree bpduguard (disable|enable)" cmd:"spanning-tree bpduguard %s"`
	ServicePolicyInput    string   `reg:"service-policy input ([[:print:]]+)" cmd:"service-policy input %s"`
	ServicePolicyOutput   string   `reg:"service-policy output ([[:print:]]+)" cmd:"service-policy output %s"`
	DhcpSnoopingThrust    bool     `reg:"ip dhcp snooping trust" cmd:"ip dhcp snooping trust"`
	Vrf                   string   `reg:"ip vrf forwarding ([[:print:]]+)" cmd:"ip vrf forwarding %s"`
	Ips                   []Ip     `reg:"ip address.*"`
	IPHelperAddresses     []string `reg:"ip helper-address (\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3})" cmd:"ip helper-address %s"`
	OspfNetwork           string   `reg:"ip ospf network (broadcast|non-broadcast|point-to-multipoint|point-to-point)" cmd:"ip ospf network %s"`
}

type Ip struct {
	Ip        string `reg:"ip address (\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}) (?:\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3})"`
	Subnet    string `reg:"ip address (?:\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}) (\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3})"`
	Secondary bool   `reg:"ip address (?:\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}) (?:\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3})( secondary)(?: vrf ([\\w\\-]+))?"`
	VRF       string `reg:"ip address (?:\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}) (?:\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3})(?: secondary)?( vrf ([\\w\\-]+))"`
}

func (inter *CiscoInterface) Parse(part string) error {
	// Get identifier
	re := regexp.MustCompile(`interface ([\w\/\.\-\:]+)`)
	identifier := re.FindStringSubmatch(part)
	identifier = strings.Split(identifier[1], ".")
	inter.Identifier = identifier[0]
	if len(identifier) > 1 {
		inter.SubInterface, _ = strconv.Atoi(identifier[1])
	}

	// Get ipv4 addresses
	/*	ipRe := regexp.MustCompile(`(?m)ip address (\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}) (\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})( secondary)?( vrf ([\w\-]+))?`)
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
		}*/

	//
	// Get all the rest stuff
	//
	err := ProcessParse(part, inter)
	if err != nil {
		return err
	}

	//Check if Routed Port, Trunk or Access when no direct config is present
	if !inter.Access && !inter.Trunk {
		if !strings.Contains(part, "ip address") && !strings.Contains(part, "no switchport") {
			inter.Access = true
		}
	}

	return nil
}

func (inter CiscoInterface) Generate() (string, error) {
	var config string
	if inter.Identifier == "" {
		return "", errors.New("missing require data for interface")
	}

	if inter.SubInterface != 0 {
		config = fmt.Sprintf("interface %s.%d\n", inter.Identifier, inter.SubInterface)
	} else {
		config = fmt.Sprintf("interface %s\n", inter.Identifier)
	}

	t := reflect.TypeOf(inter)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("cmd")
		if tag != "" {
			var cmd string
			switch field.Type.Kind() {
			case reflect.String:
				value := reflect.ValueOf(&inter).Elem().Field(i).String()
				if value == "" {
					continue
				}
				cmd = fmt.Sprintf(tag, value)
			case reflect.Int:
				value := reflect.ValueOf(&inter).Elem().Field(i).Int()
				if value == 0 {
					continue
				}
				cmd = fmt.Sprintf(tag, value)
			case reflect.Bool:
				value := reflect.ValueOf(&inter).Elem().Field(i).Bool()
				if !value {
					continue
				}
				cmd = tag
			case reflect.Float64:
				value := reflect.ValueOf(&inter).Elem().Field(i).Float()
				if value == 0.0 {
					continue
				}
				cmd = fmt.Sprintf(tag, value)
			default:
				continue
			}
			cmd = "  " + cmd + "\n"
			config = config + cmd
		}
	}
	config = config + "!"

	return config, nil
}
