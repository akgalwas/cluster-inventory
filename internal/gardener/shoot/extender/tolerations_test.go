package extender

import (
	gardener "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"

	imv1 "github.com/kyma-project/infrastructure-manager/api/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTolerationsExtender(t *testing.T) {
	for _, testCase := range []struct {
		name                string
		region              string
		expectedTolerations []gardener.Toleration
	}{
		{
			name:   "Should extend Tolerations for me-central2",
			region: "me-central2",
			expectedTolerations: []gardener.Toleration{
				{
					Key: "ksa-assured-workload",
				},
			},
		},
		{
			name:                "Should not extend Tolerations for eu-de-1",
			region:              "eu-de-1",
			expectedTolerations: nil,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			// given
			shoot := fixEmptyGardenerShoot("shoot", "kcp-system")
			basicRuntime := imv1.Runtime{
				ObjectMeta: v1.ObjectMeta{
					Name:      "runtime",
					Namespace: "namespace",
				},
				Spec: imv1.RuntimeSpec{
					Shoot: imv1.RuntimeShoot{
						Region: testCase.region,
					},
				},
			}

			// when
			err := ExtendWithTolerations(basicRuntime, &shoot)
			require.NoError(t, err)

			// then
			assert.Equal(t, testCase.expectedTolerations, shoot.Spec.Tolerations)
		})
	}
}

////
////func TestTolerationsExtender(t *testing.T) {
////	for _, testCase := range []struct {
//		name                string
//		region              string
//		expectedTolerations []gardener.Toleration
//	}{
//		{
//			name:   "Extend Tolerations for me-central2",
//
//			region: "me-central2",
//			expectedTolerations: []gardener.Toleration{
//				{
//					Key: "ksa-assured-workload",
//				},
//			},
//		},
//		{
//			name:   "Extend Tolerations for eu-central",
//			region: "me-central2",
//			expectedTolerations: []gardener.Toleration{},
//		},
//	} {
//		t.Run(testCase.name, func(t *testing.T) {
//			// given
//			shoot := fixEmptyGardenerShoot("shoot", "kcp-system")
//			basicRuntime := imv1.Runtime{
//				ObjectMeta: v1.ObjectMeta{
//					Name:      "runtime",
//					Namespace: "namespace",
//				},
//				Spec: imv1.RuntimeSpec{
//					Shoot: imv1.RuntimeShoot{
//						Region: testCase.region,
//					},
//				},
//			}
//
//			// when
//			err := ExtendWithTolerations(basicRuntime, &shoot)
//			require.NoError(t, err)
//
//			// then
//			assert.Equal(t, testCase.expectedTolerations, shoot.Spec.Tolerations)
//		})
//	}
//}
