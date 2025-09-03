package runtime

import (
	tenantruntime "nexus/admin/api/runtime/tenant"

	"nexus/base/nexus"
)

// Runtime tree.
type Runtime struct {
	nexus.SingletonNode

	// Tenant runtime spec.
	Tenant tenantruntime.Tenant `nexus:"children"  json:"tenant,omitempty"`
}
