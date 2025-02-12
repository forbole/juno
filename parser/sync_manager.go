package parser

import (
	"fmt"
	"math"
	"sync"

	"github.com/forbole/juno/v6/database"
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

func NewModuleSyncManager(db database.Database) *ModuleSyncManager {
	return &ModuleSyncManager{
		modules: make(map[string]*ModuleStatus),
		db:      db,
	}
}

func (m *ModuleSyncManager) InitModule(moduleName string) error {
	// 检查数据库中是否已有该模块的同步记录
	lastHeight, status, err := m.db.GetModuleSyncStatus(moduleName)
	if err != nil {
		return err
	}

	if lastHeight > 0 {
		// 如果数据库中有同步记录，将状态设置为 synced
		return m.db.UpdateModuleSyncStatus(moduleName, lastHeight, status)
	}

	// 完全新的模块，设置为 syncing
	return m.db.UpdateModuleSyncStatus(moduleName, 0, "syncing")
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
func (m *ModuleSyncManager) GetMinSyncedHeight() (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	minHeight := int64(math.MaxInt64)
	for _, mod := range m.modules {
		if mod.Status == "synced" && mod.LastHeight < minHeight {
			minHeight = mod.LastHeight
		}
	}
	return minHeight, nil
}

// UpdateStatusAndFlush updates a module's status and immediately writes to database
func (m *ModuleSyncManager) UpdateStatusAndFlush(name string, height int64, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if mod, exists := m.modules[name]; exists {
		mod.LastHeight = height
		mod.Status = status
		return m.db.UpdateModuleSyncStatus(name, height, status)
	}
	return fmt.Errorf("module %s not found", name)
}
