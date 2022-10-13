package jap

import (
	"io/ioutil"
	"testing"
)

type TestFile struct {
	FileName           string
	Hostname           string
	NumberOfInterfaces int
	NumberOfVlans      int
}

func TestParse(t *testing.T) {
	var testFiles []TestFile

	ciscoExampleFile := TestFile{
		FileName:           "sample.txt",
		Hostname:           "retail",
		NumberOfInterfaces: 16,
		NumberOfVlans:      5,
	}
	testFiles = append(testFiles, ciscoExampleFile)

	//switch
	genericSwitch := TestFile{
		FileName:           "genericSwitch.txt",
		Hostname:           "switch-192-168-0-50",
		NumberOfInterfaces: 16,
		NumberOfVlans:      3,
	}
	testFiles = append(testFiles, genericSwitch)

	//Router
	genericRouter := TestFile{
		FileName:           "genericRouter.txt",
		Hostname:           "router-01",
		NumberOfInterfaces: 124,
		NumberOfVlans:      65,
	}
	testFiles = append(testFiles, genericRouter)

	if len(testFiles) == 0 {
		t.Error("No test files specified")
	}
	for _, testFile := range testFiles {
		// Load the content of running config
		content, err := ioutil.ReadFile("testFiles/" + testFile.FileName)
		if err != nil {
			t.Error(err)
		}
		running, err := Parse(string(content))
		if err != nil {
			t.Error(err)
		}
		if running.FullConfig != string(content) {
			t.Error("Error loading the config")
		}

		// Check interfaces
		if len(running.Interfaces) != testFile.NumberOfInterfaces {
			t.Errorf("Wrong interface number wants: %v got %v", testFile.NumberOfInterfaces, len(running.Interfaces))
		}

		// Check Hostname
		if running.Hostname != testFile.Hostname {
			t.Errorf("Wrong hostname wants: %s got: %s", testFile.Hostname, running.Hostname)
		}

		//Check Vlans
		if len(running.Vlans) != testFile.NumberOfVlans {
			t.Errorf("Wrong vlan number wants %v got %v", testFile.NumberOfVlans, len(running.Vlans))
		}
	}

}
