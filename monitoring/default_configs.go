package monitoring

func GetDefaultThresholdValues() (plugins []Plugin) {
	return []Plugin{
		{
			Name:   "df",
			Enable: true,
			Configs: []PluginConfig{
				{Category: THRESHOLD, Type: CRITICAL, Value: "90"},
				{Category: THRESHOLD, Type: WARNING, Value: "80"},
			},
		},
		{
			Name:   "memory",
			Enable: true,
			Configs: []PluginConfig{
				{Category: THRESHOLD, Type: CRITICAL, Value: "90"},
				{Category: THRESHOLD, Type: WARNING, Value: "80"},
			},
		},
		{
			Name:   "cpu",
			Enable: true,
			Configs: []PluginConfig{
				{Category: THRESHOLD, Type: CRITICAL, Value: "90"},
				{Category: THRESHOLD, Type: WARNING, Value: "80"},
			},
		},
		{
			Name:   "swap",
			Enable: true,
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
			Name:   STORAGE_PROFILE_UTILIZATION,
			Enable: true,
			Configs: []PluginConfig{
				{Category: THRESHOLD, Type: CRITICAL, Value: "85"},
				{Category: THRESHOLD, Type: WARNING, Value: "65"},
			},
		},
	}
}

var DefaultClusterMonitoringInterval = 600
