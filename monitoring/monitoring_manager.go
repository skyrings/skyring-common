package monitoring

import (
	"fmt"
	"github.com/skyrings/skyring-common/tools/logger"
	"io"
	"os"
	"sync"
)

type MonitoringManagersFactory func(config io.Reader) (MonitoringManagerInterface, error)

var GeneralResources = []string{
	"df",
	MEMORY,
	"cpu",
	"disk",
	"network",
	"swap",
	CLUSTER_UTILIZATION,
	SYSTEM_UTILIZATION,
	STORAGE_PROFILE_UTILIZATION,
	SLU_UTILIZATION,
	NO_OF_OBJECT,
	PG_SUMMARY,
	"ping",
	"interface",
	STORAGE_UTILIZATION,
	fmt.Sprintf("aggregation-%s-%s", MEMORY, SUM),
	fmt.Sprintf("aggregation-%s-%s", SWAP, SUM),
}

var (
	monitoringManagersMutex sync.Mutex
	monitoringManagers      = make(map[string]MonitoringManagersFactory)
)

func RegisterMonitoringManager(name string, factory MonitoringManagersFactory) {
	monitoringManagersMutex.Lock()
	defer monitoringManagersMutex.Unlock()

	if _, found := monitoringManagers[name]; found {
		logger.Get().Info("Monitoring manager already registered. Returing..")
		return
	}
	monitoringManagers[name] = factory
}

func GetMonitoringManager(name string, config io.Reader) (MonitoringManagerInterface, error) {
	monitoringManagersMutex.Lock()
	defer monitoringManagersMutex.Unlock()

	factory_func, found := monitoringManagers[name]
	if !found {
		logger.Get().Info("Monitoring manager not found", name)
		return nil, nil
	}

	return factory_func(config)
}

func InitMonitoringManager(name string, configPath string) (MonitoringManagerInterface, error) {
	var manager MonitoringManagerInterface

	if name == "" {
		return nil, nil
	}

	var err error
	if configPath != "" {
		config, err := os.Open(configPath)
		if err != nil {
			logger.Get().Info("Couldnt open monitoring manager config file", configPath)
			return nil, nil
		}

		defer config.Close()
		manager, err = GetMonitoringManager(name, config)
	} else {
		manager, err = GetMonitoringManager(name, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("Could not initialize monitoring manager %s: %v", name, err)
	}
	if manager == nil {
		return nil, fmt.Errorf("Unknown monitoring manager %s", name)
	}

	return manager, nil
}
