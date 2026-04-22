package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go-micro/pkg/errx"
)

// Role defines the possible user roles in the system
type Role string

const (
	RoleAdmin          Role = "admin"
	RoleOperator       Role = "operator"
	RoleViewer         Role = "viewer"
	RoleOrderManager   Role = "order_manager"
	RolePaymentManager Role = "payment_manager"
	RoleInventoryMgr   Role = "inventory_manager"
)

// Permission defines granular permissions
type Permission string

const (
	PermOrderRead   Permission = "order:read"
	PermOrderWrite  Permission = "order:write"
	PermOrderCancel Permission = "order:cancel"

	PermPaymentRead   Permission = "payment:read"
	PermPaymentWrite  Permission = "payment:write"
	PermPaymentRefund Permission = "payment:refund"

	PermInventoryRead   Permission = "inventory:read"
	PermInventoryWrite  Permission = "inventory:write"
	PermInventoryReserve Permission = "inventory:reserve"

	PermUserRead  Permission = "user:read"
	PermUserWrite Permission = "user:write"

	PermRefundRead  Permission = "refund:read"
	PermRefundWrite Permission = "refund:write"

	PermActivityRead  Permission = "activity:read"
	PermActivityWrite Permission = "activity:write"

	PermPriceRead  Permission = "price:read"
	PermPriceWrite Permission = "price:write"

	PermDashboardRead Permission = "dashboard:read"
	PermAuditRead     Permission = "audit:read"
)

// rolePermissions maps roles to their allowed permissions
var rolePermissions = map[Role][]Permission{
	RoleAdmin: {
		PermOrderRead, PermOrderWrite, PermOrderCancel,
		PermPaymentRead, PermPaymentWrite, PermPaymentRefund,
		PermInventoryRead, PermInventoryWrite, PermInventoryReserve,
		PermUserRead, PermUserWrite,
		PermRefundRead, PermRefundWrite,
		PermActivityRead, PermActivityWrite,
		PermPriceRead, PermPriceWrite,
		PermDashboardRead, PermAuditRead,
	},
	RoleOrderManager: {
		PermOrderRead, PermOrderWrite, PermOrderCancel,
		PermDashboardRead,
	},
	RolePaymentManager: {
		PermPaymentRead, PermPaymentWrite, PermPaymentRefund,
		PermRefundRead, PermRefundWrite,
		PermDashboardRead,
	},
	RoleInventoryMgr: {
		PermInventoryRead, PermInventoryWrite, PermInventoryReserve,
		PermDashboardRead,
	},
	RoleOperator: {
		PermOrderRead, PermPaymentRead, PermInventoryRead,
		PermRefundRead, PermActivityRead, PermPriceRead,
		PermDashboardRead,
	},
	RoleViewer: {
		PermOrderRead, PermPaymentRead, PermInventoryRead,
		PermDashboardRead,
	},
}

// RequireRole returns a middleware that checks if user has one of the required roles
func RequireRole(roles ...Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get(CtxRole)
		if !exists {
			c.AbortWithStatusJSON(403, gin.H{
				"code":    errx.CodeForbidden,
				"message": "access denied: no role found",
			})
			return
		}

		userRoleStr, ok := userRole.(string)
		if !ok {
			c.AbortWithStatusJSON(403, gin.H{
				"code":    errx.CodeForbidden,
				"message": "access denied: invalid role",
			})
			return
		}

		for _, role := range roles {
			if Role(userRoleStr) == role {
				c.Next()
				return
			}
		}

		// Admin always has access
		if Role(userRoleStr) == RoleAdmin {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(403, gin.H{
			"code":    errx.CodeForbidden,
			"message": "access denied: insufficient permissions",
		})
	}
}

// RequirePermission returns a middleware that checks for specific permissions
func RequirePermission(perms ...Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get(CtxRole)
		if !exists {
			c.AbortWithStatusJSON(403, gin.H{
				"code":    errx.CodeForbidden,
				"message": "access denied: no role found",
			})
			return
		}

		userRoleStr, ok := userRole.(string)
		if !ok {
			c.AbortWithStatusJSON(403, gin.H{
				"code":    errx.CodeForbidden,
				"message": "access denied: invalid role",
			})
			return
		}

		role := Role(userRoleStr)

		// Admin has all permissions
		if role == RoleAdmin {
			c.Next()
			return
		}

		// Check if user role has any of the required permissions
		allowedPerms, ok := rolePermissions[role]
		if !ok {
			c.AbortWithStatusJSON(403, gin.H{
				"code":    errx.CodeForbidden,
				"message": "access denied: role not found",
			})
			return
		}

		for _, required := range perms {
			hasPermission := false
			for _, allowed := range allowedPerms {
				if allowed == required {
					hasPermission = true
					break
				}
			}
			if !hasPermission {
				c.AbortWithStatusJSON(403, gin.H{
					"code":    errx.CodeForbidden,
					"message": fmt.Sprintf("access denied: missing permission %s", required),
				})
				return
			}
		}

		c.Next()
	}
}

// GetUserPermissions returns all permissions for a given role
func GetUserPermissions(role Role) []Permission {
	if perms, ok := rolePermissions[role]; ok {
		return perms
	}
	return nil
}

// HasPermission checks if a role has a specific permission
func HasPermission(role Role, perm Permission) bool {
	perms := GetUserPermissions(role)
	for _, p := range perms {
		if p == perm {
			return true
		}
	}
	return false
}
