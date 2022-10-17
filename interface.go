package jap

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type CiscoInterface struct {
	Identifier            string
	SubInterface          int
	Ips                   []Ip
	IPHelperAddresses     []string `reg:"ip helper-address (\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3})" cmd:"ip helper-address %s"`
	TrunkAllowedVlan      []int    `reg:"ip helper-address (\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3})" cmd:"ip helper-address %s"`
	Shutdown              bool     `reg:"shutdown" cmd:"shutdown"`
	Trunk                 bool     `reg:"switchport mode trunk" cmd:"switchport mode trunk"`
	AccessPort            bool     `reg:"switchport mode access" cmd:"switchport mode access"`
	DhcpSnoopingThrust    bool     `reg:"ip dhcp snooping trust" cmd:"ip dhcp snooping trust"`
	PortSecurity          bool     `reg:"switchport port-security" cmd:"switchport port-security"`
	Description           string   `reg:"description ([[:print:]]+)" cmd:"description %s"`
	Vrf                   string   `reg:"ip vrf forwarding ([[:print:]]+)" cmd:"ip vrf forwarding %s"`
	PortSecurityAgingType string   `reg:"switchport port-security type (absolute|inactivity)" cmd:"switchport port-security type  %s"`
	PortSecurityViolation string   `reg:"switchport port-security violation (protect|restrict|shutdown)" cmd:"switchport port-security violation %s"`
	STPportfast           string   `reg:"spanning-tree portfast (disable|edge|network)" cmd:"spanning-tree portfast %s"`
	STPbpduguard          string   `reg:"spanning-tree bpduguard (disable|enable)" cmd:"spanning-tree bpduguard %s"`
	NativeVlan            int      `reg:"switchport trunk native vlan ([0-9]+)" cmd:"switchport trunk native vlan %d"`
	AccessVlan            int      `reg:"switchport access vlan ([0-9]+)" cmd:"switchport access vlan %d"`
	VoiceVlan             int      `reg:"switchport voice vlan ([0-9]+)" cmd:"switchport voice vlan %d"`
	PortSecurityMaximum   int      `reg:"switchport port-security maximum ([0-9]+)" cmd:"switchport port-security maximum %d"`
	PortSecurityAgingTime int      `reg:"switchport port-security aging time ([0-9]+)" cmd:"switchport port-security aging time %d"`
	SCBroadcastLevel      float64  `reg:"storm-control broadcast level ([0-9.]+)" cmd:"storm-control broadcast level %f"`
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
		tag := field.Tag.Get("reg")
		if tag != "" {
			re = regexp.MustCompile(tag)
			// @todo check if no is with the command!
			data := re.FindStringSubmatch(part)
			if len(data) == 0 {
				continue
			}

			switch field.Type.Kind() {
			case reflect.String:
				reflect.ValueOf(&inter).Elem().Field(i).SetString(data[1])
				break
			case reflect.Int:
				value, err := strconv.ParseInt(data[1], 10, 64)
				if err != nil {
					return CiscoInterface{}, err
				}
				reflect.ValueOf(&inter).Elem().Field(i).SetInt(value)
				break
			case reflect.Bool:
				reflect.ValueOf(&inter).Elem().Field(i).SetBool(true)
				break
			case reflect.Float64:
				float, err := strconv.ParseFloat(data[1], 64)
				if err != nil {
					return CiscoInterface{}, err
				}
				reflect.ValueOf(&inter).Elem().Field(i).SetFloat(float)
			}
		}
	}

	//Check if Routed Port, Trunk or Access when no direct config is present
	if !inter.AccessPort && !inter.Trunk {
		if !strings.Contains(part, "ip address") && !strings.Contains(part, "no switchport") {
			inter.AccessPort = true
		}
	}

	return inter, nil
}

func (in CiscoInterface) GenerateInterface() (string, error) {
	var config string
	if in.Identifier == "" {
		return "", errors.New("missing require data for interface")
	}

	if in.SubInterface != 0 {
		config = fmt.Sprintf("interface %s.%d\n", in.Identifier, in.SubInterface)
	} else {
		config = fmt.Sprintf("interface %s\n", in.Identifier)
	}

	t := reflect.TypeOf(in)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("cmd")
		if tag != "" {
			var cmd string
			switch field.Type.Kind() {
			case reflect.String:
				value := reflect.ValueOf(&in).Elem().Field(i).String()
				if value == "" {
					continue
				}
				cmd = fmt.Sprintf(tag, value)
			case reflect.Int:
				value := reflect.ValueOf(&in).Elem().Field(i).Int()
				if value == 0 {
					continue
				}
				cmd = fmt.Sprintf(tag, value)
			case reflect.Bool:
				value := reflect.ValueOf(&in).Elem().Field(i).Bool()
				if !value {
					continue
				}
				cmd = tag
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
