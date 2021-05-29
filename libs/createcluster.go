package libs

import (
	"context"
	hivev1 "github.com/openshift/hive/apis/hive/v1"
	"github.com/openshift/hive/apis/hive/v1/aws"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

func CreateClusterDeployment(labRequest *LabRequest) {
	cfg, err := DefaultClientK8sAuthenticate()
	ErrorCheck("Unable to create default client: %v\n", err)

	scheme := runtime.NewScheme()
	err = hivev1.SchemeBuilder.AddToScheme(scheme)
	ErrorCheck("Unable to add hive to scheme: %v\n", err)

	dc, err := client.New(cfg, client.Options{Scheme: scheme})
	ErrorCheck("Unable to create K8s client: %v\n", err)

	kc := K8sAuthenticate()
	labSecret, err := kc.CoreV1().Secrets("hive").Get(context.Background(), labRequest.ID.String(), metav1.GetOptions{})
	ErrorCheck("Unable to get the lab secret; does it exist: ", err)

	// TODO: #1 Allow selection of platform; will require some Google Form changes and potentially capturing
	// information from partner specific to the platform cluster should be installed on
	// using AWS for now
	plat := hivev1.Platform{
		AWS:            &aws.Platform{
			CredentialsSecretRef: corev1.LocalObjectReference{Name: "hive-aws-creds"},
			Region:               "us-west-2", // It could be useful to allow region selection based on where partner is
			UserTags:             map[string]string{"LabID": labRequest.ID.String()},
		},
	}

	secretRef := corev1.LocalObjectReference{Name: labRequest.ID.String()}

	leasetime := []string{"one-day", "one-week", "two-weeks", "one-month"}

	oplLabels := map[string]string{
		"opl-region": labRequest.Availability,
		"opl-lease-time": leasetime[labRequest.LeaseTime],
	}

	charsFromID := strings.Split(labRequest.ID.String(), "-")[0]
	cds := hivev1.ClusterDeploymentSpec{
		ClusterName:          labRequest.ClusterName + "-" + charsFromID,
		BaseDomain:           "opdev.io",
		Platform:             plat,
		ManageDNS:            false,
		Provisioning:         &hivev1.Provisioning{
			InstallConfigSecretRef: &secretRef,
			ImageSetRef:            &hivev1.ClusterImageSetReference{Name: string(labSecret.Data["openshift"])},
			SSHPrivateKeySecretRef: &secretRef,
		},
	}

	cd := hivev1.ClusterDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:                       labRequest.ID.String(),
			Namespace:                  "hive",
			Labels:											oplLabels,
		},
		Spec:       cds,
	}

	err = dc.Create(context.Background(), &cd)
	ErrorCheck("Unable to create cluster deployment: ", err)
}


