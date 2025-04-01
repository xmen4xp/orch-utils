// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package kubernetes

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/open-edge-platform/orch-utils/secrets"
)

type StorageService struct {
	client *kubernetes.Clientset
}

var _ secrets.StorageService = &StorageService{}

func NewStorageService(cli *kubernetes.Clientset) (*StorageService, error) {
	return &StorageService{
		client: cli,
	}, nil
}

// Put stores value at name in the data store.
func (svc *StorageService) Put(ctx context.Context, namespace string, name string, values map[string]string) error {
	if _, err := svc.client.CoreV1().Secrets(namespace).Create(
		ctx,
		&v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			StringData: values,
		},
		metav1.CreateOptions{},
	); err != nil {
		return fmt.Errorf("create secret: %w", err)
	}

	return nil
}

// Get retrieves values at name in the data store.
func (svc *StorageService) Get(ctx context.Context, namespace string, name string) (map[string]string, error) {
	secret, err := svc.client.CoreV1().Secrets(namespace).Get(
		ctx,
		name,
		metav1.GetOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("get secret %s: %w", name, err)
	}

	values := map[string]string{}
	for k, v := range secret.Data {
		values[k] = string(v)
	}

	return values, nil
}

// Delete by name in the data store.
func (svc *StorageService) Delete(ctx context.Context, namespace string, name string) error {
	if err := svc.client.CoreV1().Secrets(namespace).Delete(
		ctx,
		name,
		metav1.DeleteOptions{},
	); err != nil {
		return fmt.Errorf("delete secret %s: %w", name, err)
	}

	return nil
}
