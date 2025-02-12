package parser

import (
	"fmt"
	"math"
	"sync"

	"github.com/forbole/juno/v6/logging"
	"github.com/forbole/juno/v6/node"

	"github.com/forbole/juno/v6/database"
	"github.com/forbole/juno/v6/modules"
)

// ModuleSyncManager manages the sync status of modules
type ModuleSyncManager struct {
	mu      sync.RWMutex
	modules map[string]*ModuleStatus
	db      database.Database
}

type ModuleStatus struct {
	Name       string
	LastHeight int64
	Status     string // "syncing" or "synced"
}

// Context represents the context that is shared among different workers
type Context struct {
	Node        node.Node
	Database    database.Database
	Logger      logging.Logger
	Modules     []modules.Module
	SyncManager *ModuleSyncManager
}

// NewContext builds a new Context instance
func NewContext(
	proxy node.Node, db database.Database,
	logger logging.Logger, modules []modules.Module,
) *Context {
	return &Context{
		Node:        proxy,
		Database:    db,
		Modules:     modules,
		Logger:      logger,
		SyncManager: NewModuleSyncManager(db),
	}
}

func NewModuleSyncManager(db database.Database) *ModuleSyncManager {
	return &ModuleSyncManager{
		modules: make(map[string]*ModuleStatus),
		db:      db,
	}
}

func (m *ModuleSyncManager) InitModule(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Try to load existing status
	lastHeight, status, err := m.db.GetModuleSyncStatus(name)
	if err != nil {
		// If not found, initialize with syncing status
		status = "syncing"
		lastHeight = 0
		if err := m.db.SaveModuleSyncStatus(name, lastHeight, status); err != nil {
			return err
		}
	}

	m.modules[name] = &ModuleStatus{
		Name:       name,
		LastHeight: lastHeight,
		Status:     status,
	}
	return nil
}

func (m *ModuleSyncManager) UpdateStatus(name string, height int64, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.db.UpdateModuleSyncStatus(name, height, status); err != nil {
		return err
	}

	if mod, exists := m.modules[name]; exists {
		mod.LastHeight = height
		mod.Status = status
	}
	return nil
}

func (m *ModuleSyncManager) GetStatus(name string) (ModuleStatus, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if mod, exists := m.modules[name]; exists {
		return *mod, true
	}
	return ModuleStatus{}, false
}

// UpdateHeight updates the last processed height for a module
func (m *ModuleSyncManager) UpdateHeight(name string, height int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if mod, exists := m.modules[name]; exists {
		mod.LastHeight = height
		return nil
	}
	return fmt.Errorf("module %s not found", name)
}

// GetMinSyncedHeight returns the minimum height among all synced modules
func (m *ModuleSyncManager) GetMinSyncedHeight() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	minHeight := int64(math.MaxInt64)
	for _, mod := range m.modules {
		if mod.Status == "synced" && mod.LastHeight < minHeight {
			minHeight = mod.LastHeight
		}
	}
	return minHeight
}

// FlushToDB writes all current heights to the database
func (m *ModuleSyncManager) FlushToDB() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for name, mod := range m.modules {
		if err := m.db.UpdateModuleSyncStatus(name, mod.LastHeight, mod.Status); err != nil {
			return fmt.Errorf("failed to update module %s status: %s", name, err)
		}
	}
	return nil
}
