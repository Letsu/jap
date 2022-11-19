package jap

import (
	"io/ioutil"
	"testing"
)

type testOspf struct {
	FileName string
}

func TestOspf_Parse(t *testing.T) {
	var ospfFiles []testOspf
	process := testOspf{
		FileName: "process.txt",
	}
	ospfFiles = append(ospfFiles, process)
	processVrf := testOspf{
		FileName: "processVrf.txt",
	}
	ospfFiles = append(ospfFiles, processVrf)

	for _, ospfFile := range ospfFiles {
		// Open File
		content, err := ioutil.ReadFile("testFiles/ospf/" + ospfFile.FileName)
		if err != nil {
			t.Error(err)
		}

		var ospf Ospf
		err = ospf.Parse(string(content))
		if err != nil {
			t.Error(err)
		}

		generated, err := ospf.Generate()
		if err != nil {
			t.Error(err)
		}

		if generated != string(content) {
			t.Error("Config wrong parsed or generated \n File: \n", string(content), "\n Generated: \n", generated)
		}
	}
}
