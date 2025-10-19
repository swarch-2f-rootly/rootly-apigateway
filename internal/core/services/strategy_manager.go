package services

import (
	"context"
	"fmt"
	"sync"

	"github.com/swarch-2f-rootly/rootly-apigateway/internal/core/ports"
)

// StrategyManager implements the strategy management functionality
type StrategyManager struct {
	strategies map[string]ports.RouteStrategy
	mutex      sync.RWMutex
	logger     ports.Logger
}

// NewStrategyManager creates a new strategy manager
func NewStrategyManager(logger ports.Logger) *StrategyManager {
	return &StrategyManager{
		strategies: make(map[string]ports.RouteStrategy),
		logger:     logger,
	}
}

// RegisterStrategy registers a new strategy
func (sm *StrategyManager) RegisterStrategy(name string, strategy ports.RouteStrategy) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.strategies[name] = strategy
	sm.logger.Info("Strategy registered", map[string]interface{}{
		"strategy_name": name,
		"strategy_type": fmt.Sprintf("%T", strategy),
	})
}

// GetStrategy retrieves a strategy by name
func (sm *StrategyManager) GetStrategy(name string) (ports.RouteStrategy, bool) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	strategy, exists := sm.strategies[name]
	return strategy, exists
}

// ExecuteStrategy executes a strategy with the given parameters
func (sm *StrategyManager) ExecuteStrategy(ctx context.Context, strategyName string, params ports.StrategyParams) (interface{}, error) {
	strategy, exists := sm.GetStrategy(strategyName)
	if !exists {
		return nil, fmt.Errorf("strategy not found: %s", strategyName)
	}

	// Validar que params y Request no sean nil
	if params.Request == nil {
		return nil, fmt.Errorf("request is nil")
	}

	requestPath := ""
	if params.Request.URL != nil {
		requestPath = params.Request.URL.Path
	}

	sm.logger.Debug("Executing strategy", map[string]interface{}{
		"strategy_name": strategyName,
		"request_path":  requestPath,
	})

	result, err := strategy.Execute(ctx, params)
	if err != nil {
		sm.logger.Error("Strategy execution failed", err, map[string]interface{}{
			"strategy_name": strategyName,
			"request_path":  requestPath,
		})
		return nil, err
	}

	sm.logger.Debug("Strategy executed successfully", map[string]interface{}{
		"strategy_name": strategyName,
		"request_path":  requestPath,
	})

	return result, nil
}

// ListStrategies returns a list of all registered strategies
func (sm *StrategyManager) ListStrategies() []string {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	strategies := make([]string, 0, len(sm.strategies))
	for name := range sm.strategies {
		strategies = append(strategies, name)
	}
	return strategies
}

// UnregisterStrategy removes a strategy
func (sm *StrategyManager) UnregisterStrategy(name string) bool {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if _, exists := sm.strategies[name]; exists {
		delete(sm.strategies, name)
		sm.logger.Info("Strategy unregistered", map[string]interface{}{
			"strategy_name": name,
		})
		return true
	}
	return false
}