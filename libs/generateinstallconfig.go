package libs

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/template"
)

func GenerateInstallConfig(labRequest *LabRequest) []byte {
	// TODO: #1 Align LabRequest options more with InstallConfig options
	//InstallConfig{
	//	BaseDomain:        "",
	//	WorkerReplicas:    0,
	//	MasterReplicas:    0,
	//	ClusterName:       "",
	//	NetworkType:       "",
	//	ServiceNetwork:    "",
	//	Cloud:             "",
	//	RegionDesignation: "",
	//	Region:            "",
	//	PullSecret:        "",
	//	PublicSSHKey:      "",
	//}
	charsFromID := strings.Split(labRequest.ID.String(), "-")[0]

	clustersizes := map[int][]string{
		0: {"m5.xlarge", "m5.large"},
		1: {"m5.xlarge", "m5.xlarge"},
		2: {"m5.xlarge", "m5.2xlarge"},
	}

	ic := InstallConfig{
		ClusterName:       labRequest.ClusterName + "-" + charsFromID,
		PublicSSHKey:      labRequest.PublicSSHKey,
		MasterSize:        clustersizes[labRequest.ClusterSize][0],
		WorkerSize:        clustersizes[labRequest.ClusterSize][1],
	}

	tmpfile := "/tmp/" + labRequest.ID.String() + ".ic"

	paths := []string{
		"/tmp/install-config.tmpl",
	}

	t, err := template.New("install-config.tmpl").ParseFiles(paths...)
	if err != nil {
		log.Fatalf("Unable to parse template: %v\n", err)
	}

	// TODO: #2 Explore options to not create file
	icfile, err := os.OpenFile(tmpfile, os.O_RDWR|os.O_CREATE, 0755)
	ErrorCheck("Unable to open lab request tmp install-config: ", err)

	err = t.Execute(icfile, ic)
	if err != nil {
		log.Fatalf("Unable to construct template: %v\n", err)
	}

	if err = icfile.Close(); err != nil {
		log.Fatalf("FATALITY! Unable to close lab request tmp install-config: %v", err)
	}

	data, err := ioutil.ReadFile(tmpfile)
	ErrorCheck("Unable to read lab request tmp install-config: ", err)

	return data
}