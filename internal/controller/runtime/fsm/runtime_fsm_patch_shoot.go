package fsm

import (
	"context"
	"slices"

	gardener "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	imv1 "github.com/kyma-project/infrastructure-manager/api/v1"
	gardener_shoot "github.com/kyma-project/infrastructure-manager/pkg/gardener/shoot"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var fieldManagerName = "kim"

func sFnPatchExistingShoot(ctx context.Context, m *fsm, s *systemState) (stateFn, *ctrl.Result, error) {
	m.log.Info("Patch shoot state")

	fetchAuditLogConfig := buildFetchAuditLogConfig(get(m.ShootClient.Get), m.AuditLogging)
	data, err := fetchAuditLogConfig(ctx, s.shoot.Spec.SeedName, s.shoot.Spec.Region)

	if err != nil {
		m.log.Error(err, "Failed to configure audit logs", "auditLogMandatory", m.AuditLogMandatory)
	}

	if err != nil && m.AuditLogMandatory {
		s.instance.UpdateStatePending(
			imv1.ConditionTypeRuntimeProvisioned,
			imv1.ConditionReasonAuditLogError,
			"False",
			err.Error(),
		)
		return updateStatusAndStop()
	}

	zonesFromShoot := getZones(s.shoot.Spec.Provider.Workers)
	updatedShoot, err := convertPatch(&s.instance, gardener_shoot.PatchOpts{
		ConverterConfig: m.ConverterConfig,
		AuditLogData:    data,
		Zones:           zonesFromShoot,
	})

	if err != nil {
		m.log.Error(err, "Failed to convert Runtime instance to shoot object, exiting with no retry")
		m.Metrics.IncRuntimeFSMStopCounter()
		return updateStatePendingWithErrorAndStop(&s.instance, imv1.ConditionTypeRuntimeProvisioned, imv1.ConditionReasonConversionError, "Runtime conversion error")
	}

	m.log.Info("Shoot converted successfully", "Name", updatedShoot.Name, "Namespace", updatedShoot.Namespace)

	err = m.ShootClient.Patch(ctx, &updatedShoot, client.Apply, &client.PatchOptions{
		FieldManager: fieldManagerName,
		Force:        ptr.To(true),
	})

	if err != nil {
		if k8serrors.IsConflict(err) {
			m.log.Info("Gardener shoot for runtime is outdated, retrying", "Name", s.shoot.Name, "Namespace", s.shoot.Namespace)
			return updateStatusAndRequeueAfter(m.RCCfg.GardenerRequeueDuration)
		}

		m.log.Error(err, "Failed to patch shoot object, exiting with no retry")
		m.Metrics.IncRuntimeFSMStopCounter()
		return updateStatePendingWithErrorAndStop(&s.instance, imv1.ConditionTypeRuntimeProvisioned, imv1.ConditionReasonProcessingErr, "Shoot patch error")
	}

	if updatedShoot.Generation == s.shoot.Generation {
		m.log.Info("Gardener shoot for runtime did not change after patch, moving to processing", "Name", s.shoot.Name, "Namespace", s.shoot.Namespace)
		return switchState(sFnConfigureOidc)
	}

	m.log.Info("Gardener shoot for runtime patched successfully", "Name", s.shoot.Name, "Namespace", s.shoot.Namespace)

	s.instance.UpdateStatePending(
		imv1.ConditionTypeRuntimeProvisioned,
		imv1.ConditionReasonProcessing,
		"Unknown",
		"Shoot is pending for update",
	)

	return updateStatusAndRequeueAfter(m.RCCfg.GardenerRequeueDuration)
}

func getZones(workers []gardener.Worker) []string {
	var zones []string

	for _, worker := range workers {
		for _, zone := range worker.Zones {
			if !slices.Contains(zones, zone) {
				zones = append(zones, zone)
			}
		}
	}

	return zones
}

func convertPatch(instance *imv1.Runtime, opts gardener_shoot.PatchOpts) (gardener.Shoot, error) {
	if err := instance.ValidateRequiredLabels(); err != nil {
		return gardener.Shoot{}, err
	}

	converter := gardener_shoot.NewConverterPatch(opts)
	newShoot, err := converter.ToShoot(*instance)
	if err != nil {
		return newShoot, err
	}

	return newShoot, nil
}

func updateStatePendingWithErrorAndStop(instance *imv1.Runtime,
	//nolint:unparam
	c imv1.RuntimeConditionType, r imv1.RuntimeConditionReason, msg string) (stateFn, *ctrl.Result, error) {
	instance.UpdateStatePending(c, r, "False", msg)
	return updateStatusAndStop()
}
