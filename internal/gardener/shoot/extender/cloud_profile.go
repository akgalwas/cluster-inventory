package extender

import (
	gardener "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	imv1 "github.com/kyma-project/infrastructure-manager/api/v1"
	"github.com/kyma-project/infrastructure-manager/internal/gardener/shoot/hyperscaler"
	"github.com/pkg/errors"
)

const (
	DefaultAWSCloudProfileName       = "aws"
	DefaultAzureCloudProfileName     = "az"
	DefaultGCPCloudProfileName       = "gcp"
	DefaultOpenStackCloudProfileName = "converged-cloud-kyma"
)

func ExtendWithCloudProfile(runtime imv1.Runtime, shoot *gardener.Shoot) error {
	cloudProfileName, err := getCloudProfileName(runtime)

	if err != nil {
		return err
	}

	shoot.Spec.CloudProfileName = cloudProfileName

	return nil
}

func getCloudProfileName(runtime imv1.Runtime) (string, error) {
	switch runtime.Spec.Shoot.Provider.Type {
	case hyperscaler.TypeAWS:
		return DefaultAWSCloudProfileName, nil
	case hyperscaler.TypeGCP:
		return DefaultGCPCloudProfileName, nil
	case hyperscaler.TypeAzure:
		return DefaultAzureCloudProfileName, nil
	case hyperscaler.TypeOpenStack:
		return DefaultOpenStackCloudProfileName, nil
	}

	return "", errors.New("provider not supported")
}
