// Package types contains shared types that lack a better home.
package types

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/telekom/cluster-api-ipam-provider-in-cluster/api/v1alpha2"
)

// GenericInClusterPool is a common interface for InClusterIPPool and GlobalInClusterIPPool.
type GenericInClusterPool interface {
	client.Object
	PoolSpec() *v1alpha2.InClusterIPPoolSpec
}
