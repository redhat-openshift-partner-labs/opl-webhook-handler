module github.com/rhecoeng/opl-webhook-handler

go 1.16

require (
	github.com/go-playground/validator v9.31.0+incompatible
	github.com/google/go-github/v33 v33.0.0
	github.com/google/uuid v1.2.0
	github.com/openshift/hive/apis v0.0.0-20210528032741-c6db6f1aa0ae
	github.com/rhecoeng/utils v0.0.0-20210529051434-0866892a21d0
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a
	golang.org/x/oauth2 v0.0.0-20210514164344-f6687ab2804c
	gopkg.in/go-playground/webhooks.v5 v5.17.0
	k8s.io/api v0.21.1
	k8s.io/apimachinery v0.21.1
	k8s.io/client-go v0.21.1
	sigs.k8s.io/controller-runtime v0.8.3
)
