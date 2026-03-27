// Package authz provides a Casbin-backed authorization layer.
// Use New to create a CasbinAuthorizer, SetDefault to register it globally,
// and Check for per-request permission evaluation with automatic policy reload.
package authz

import (
	"errors"
	"sync"

	"github.com/casbin/casbin/v3"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

// Authorizer is the permission-check interface used by the Authz middleware.
type Authorizer interface {
	Check(sub, obj, act string) (bool, error)
}

type Reloader interface {
	Reload() error
}

type CasbinAuthorizer struct {
	enforcer *casbin.Enforcer
}

var (
	defaultMu         sync.RWMutex
	defaultAuthorizer Authorizer = noopAuthorizer{}
)

type noopAuthorizer struct{}

func (noopAuthorizer) Check(_, _, _ string) (bool, error) {
	return false, errors.New("authorizer not configured")
}

// New creates a CasbinAuthorizer backed by the given GORM database and
// Casbin model file. Policy is loaded from the database on construction.
func New(db *gorm.DB, modelPath string) (*CasbinAuthorizer, error) {
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, err
	}

	enforcer, err := casbin.NewEnforcer(modelPath, adapter)
	if err != nil {
		return nil, err
	}
	if err := enforcer.LoadPolicy(); err != nil {
		return nil, err
	}

	return &CasbinAuthorizer{enforcer: enforcer}, nil
}

func (a *CasbinAuthorizer) Check(sub, obj, act string) (bool, error) {
	return a.enforcer.Enforce(sub, obj, act)
}

func (a *CasbinAuthorizer) Reload() error {
	return a.enforcer.LoadPolicy()
}

func (a *CasbinAuthorizer) Enforcer() *casbin.Enforcer {
	return a.enforcer
}

// SetDefault registers a global Authorizer used by Check.
// Pass nil to reset to the no-op authorizer.
func SetDefault(authorizer Authorizer) {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	if authorizer == nil {
		defaultAuthorizer = noopAuthorizer{}
		return
	}
	defaultAuthorizer = authorizer
}

// Check evaluates whether subject sub may perform act on obj using the
// global Authorizer. Policy is reloaded on every call when the authorizer
// implements Reloader, ensuring rule changes take effect immediately.
func Check(sub, obj, act string) (bool, error) {
	defaultMu.RLock()
	authorizer := defaultAuthorizer
	defaultMu.RUnlock()

	if reloader, ok := authorizer.(Reloader); ok {
		if err := reloader.Reload(); err != nil {
			return false, err
		}
	}

	return authorizer.Check(sub, obj, act)
}
