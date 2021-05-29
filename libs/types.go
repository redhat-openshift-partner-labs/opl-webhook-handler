package libs

import "github.com/google/uuid"

// LabRequest is data captured from filling out the lab request Google Form.
type LabRequest struct {
	Timestamp                    string    `json:"time"`
	Epoch                        int       `json:"epoch" validate:"required"`
	ID                           uuid.UUID `json:"labid" validate:"omitempty"`
	LeaseTime                    int       `json:"leaseTime" validate:"omitempty"`
	PrimaryContactName           string    `json:"primaryContactName" validate:"required"`
	PrimaryContactEmail          string    `json:"primaryContactEmail" validate:"required,email"`
	PrimaryContactPhoneNumber    string    `json:"primaryContactPhoneNumber" validate:"omitempty"`
	PrimaryContactConnectUser    bool      `json:"isPrimaryContactConnectUser" validate:"omitempty"`
	SecondaryContactName         string    `json:"secondaryContactName" validate:"required"`
	SecondaryContactEmail        string    `json:"secondaryContactEmail" validate:"required,email"`
	SecondaryContactPhoneNumber  string    `json:"secondaryContactPhoneNumber" validate:"omitempty"`
	SecondaryContactConnectUser  bool      `json:"isSecondaryContactConnectUser" validate:"omitempty"`
	RedHatSponsor                string    `json:"redHatSponsor" validate:"required"`
	Availability                 string    `json:"availability" validate:"required"`
	CompanyName                  string    `json:"companyName" validate:"required"`
	CompanyConnectPartner        bool      `json:"isCompanyConnectPartner" validate:"omitempty"`
	CertificationProject         string    `json:"certificationProject" validate:"omitempty"`
	IntendedCertificationProject string    `json:"intendedCertificationProject" validate:"omitempty"`
	ProjectName                  string    `json:"projectName" validate:"omitempty"`
	PublicSSHKey                 string    `json:"publicsshkey" validate:"omitempty"`
	ClusterName                  string    `json:"clusterName" validate:"required"`
	ClusterSize                  int       `json:"clusterSize" validate:"omitempty"`
	OpenShiftVersion             string    `json:"openShiftVersion" validate:"required"`
	Description                  string    `json:"description" validate:"omitempty"`
	Notes                        string    `json:"notes" validate:"omitempty"`
}

type LabRequests struct {
	Labs []LabRequest `json:"labs"`
}

type FormRequest struct {
	Title string `json:"title" validate:"required"`
	Body  string `json:"body" validate:"required"`
}

// LabRequestBranch is the branch created when a LabRequest has been validated
// and approved. This branch is used when creating a PR for the LabRequest and
// is based on latest master
type LabRequestBranch struct {
	Base string `json:"base"`
	Lab  string `json:"labid"`
}

// LabRequestFile is the file generated for a pull request when a LabRequest
// has been validated and approved. This file is created prior to creating the
// pull request.
type LabRequestFile struct {
	FileName          string `json:"filename"`
	FileCommitMessage string `json:"filecommitmessage"`
	FileContent       string `json:"filecontent"`
}

type InstallConfig struct {
	BaseDomain        string
	WorkerReplicas    int
	MasterReplicas    int
	MasterSize        string
	WorkerSize        string
	ClusterName       string
	NetworkType       string
	ServiceNetwork    string
	Cloud             string
	RegionDesignation string
	Region            string
	PullSecret        string
	PublicSSHKey      string
}
