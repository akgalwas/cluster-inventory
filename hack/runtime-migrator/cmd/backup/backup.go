package main

import (
	"context"
	"fmt"
	"github.com/gardener/gardener/pkg/apis/core/v1beta1"
	gardener_types "github.com/gardener/gardener/pkg/client/core/clientset/versioned/typed/core/v1beta1"
	"github.com/kyma-project/infrastructure-manager/hack/runtime-migrator-app/internal/backup"
	"github.com/kyma-project/infrastructure-manager/hack/runtime-migrator-app/internal/config"
	"github.com/kyma-project/infrastructure-manager/pkg/gardener/kubeconfig"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log/slog"
	"time"
)

const (
	timeoutK8sOperation = 20 * time.Second
	expirationTime      = 60 * time.Minute
	runtimeIDAnnotation = "kcp.provisioner.kyma-project.io/runtime-id"
)

type Backup struct {
	shootClient        gardener_types.ShootInterface
	kubeconfigProvider kubeconfig.Provider
	outputWriter       backup.OutputWriter
	cfg                config.Config
}

func NewBackup(cfg config.Config, kubeconfigProvider kubeconfig.Provider, shootClient gardener_types.ShootInterface) (Backup, error) {
	outputWriter, err := backup.NewOutputWriter(cfg.OutputPath)
	if err != nil {
		return Backup{}, err
	}

	return Backup{
		shootClient:        shootClient,
		kubeconfigProvider: kubeconfigProvider,
		outputWriter:       outputWriter,
		cfg:                cfg,
	}, nil
}

func (b Backup) Do(ctx context.Context, runtimeIDs []string) error {
	listCtx, cancel := context.WithTimeout(ctx, timeoutK8sOperation)
	defer cancel()

	shootList, err := b.shootClient.List(listCtx, v1.ListOptions{})
	if err != nil {
		return err
	}

	backuper := backup.NewBackuper(b.cfg.IsDryRun, b.kubeconfigProvider)
	for _, runtimeID := range runtimeIDs {
		shoot, err := b.fetchShoot(ctx, shootList, runtimeID)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to fetch runtime: %v", err), "runtimeID", runtimeID)
			continue
		}

		runtimeBackup, err := backuper.Do(ctx, *shoot)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to backup runtime: %v", err), "runtimeID", runtimeID)
			continue
		}

		if !b.cfg.IsDryRun {
			if err := b.outputWriter.Save(runtimeBackup); err != nil {
				slog.Error(fmt.Sprintf("Failed to store backup: %v", err), "runtimeID", runtimeID)
			}
		}
	}

	return nil
}

func (b Backup) fetchShoot(ctx context.Context, shootList *v1beta1.ShootList, runtimeID string) (*v1beta1.Shoot, error) {
	shoot := findShoot(runtimeID, shootList)
	if shoot == nil {
		return nil, nil
	}

	getCtx, cancel := context.WithTimeout(ctx, timeoutK8sOperation)
	defer cancel()

	// We are fetching the shoot from the gardener to make sure the runtime didn't get deleted during the migration process
	refreshedShoot, err := b.shootClient.Get(getCtx, shoot.Name, v1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	return refreshedShoot, nil
}

func findShoot(runtimeID string, shootList *v1beta1.ShootList) *v1beta1.Shoot {
	for _, shoot := range shootList.Items {
		if shoot.Annotations[runtimeIDAnnotation] == runtimeID {
			return &shoot
		}
	}
	return nil
}
