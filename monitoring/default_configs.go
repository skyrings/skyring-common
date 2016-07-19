package monitoring

func GetDefaultThresholdValues() (plugins []Plugin) {
	return []Plugin{
		{
			Name:        "df",
			Description: "Mount Point",
			Enable:      true,
			Configs: []PluginConfig{
				{Category: THRESHOLD, Type: CRITICAL, Value: "90"},
				{Category: THRESHOLD, Type: WARNING, Value: "80"},
			},
		},
		{
			Name:        "memory",
			Description: "Memory",
			Enable:      true,
			Configs: []PluginConfig{
				{Category: THRESHOLD, Type: CRITICAL, Value: "90"},
				{Category: THRESHOLD, Type: WARNING, Value: "80"},
			},
		},
		{
			Name:        "cpu",
			Description: "CPU",
			Enable:      true,
			Configs: []PluginConfig{
				{Category: THRESHOLD, Type: CRITICAL, Value: "90"},
				{Category: THRESHOLD, Type: WARNING, Value: "80"},
			},
		},
		{
			Name:        "swap",
			Description: "Swap",
			Enable:      true,
			Configs: []PluginConfig{
				{Category: THRESHOLD, Type: CRITICAL, Value: "70"},
				{Category: THRESHOLD, Type: WARNING, Value: "50"},
			},
		},
	}
}

func GetSystemDefaultThresholdValues() map[string]Plugin {
	return map[string]Plugin{
		STORAGE_PROFILE_UTILIZATION: {
			Name:        STORAGE_PROFILE_UTILIZATION,
			Description: "Storage Profile",
			Enable:      true,
			Configs: []PluginConfig{
				{Category: THRESHOLD, Type: CRITICAL, Value: "85"},
				{Category: THRESHOLD, Type: WARNING, Value: "65"},
			},
		},
	}
}

var DefaultClusterMonitoringInterval = 600
