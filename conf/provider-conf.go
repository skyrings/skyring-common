/*Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package conf

import (
	"encoding/json"
	"github.com/skyrings/skyring-common/tools/logger"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"
)

type Route struct {
	Name       string `json:"name"`
	Method     string `json:"method"`
	Pattern    string `json:"pattern"`
	PluginFunc string `json:"pluginFunc"`
	Version    int    `json:"version"`
}

type ProviderConfig struct {
	Name           string  `json:"name"`
	ProviderBinary string  `json:"binary"`
	CompatVersion  float64 `json:"compatible_version"`
}

type ProviderInfo struct {
	Provider        ProviderConfig         `json:"provider"`
	Routes          []Route                `json:"routes"`
	Provisioner     ProvisionerConfig      `json:"provisioner"`
	ProviderOptions map[string]interface{} `json:"provideroptions"`
}

func LoadProviderConfig(providerConfigDir string) []ProviderInfo {
	var (
		data       ProviderInfo
		collection []ProviderInfo
	)

	files, err := ioutil.ReadDir(providerConfigDir)
	if err != nil {
		logger.Get().Error("Unable to read directory. error: %v", err)
		logger.Get().Error("Failed to initialize")
		return collection
	}
	for _, f := range files {
		logger.Get().Debug("File Name: %s", f.Name())

		file, err := ioutil.ReadFile(path.Join(providerConfigDir, f.Name()))
		if err != nil {
			logger.Get().Critical("Error reading config. error: %v", err)
			continue
		}
		if strings.HasSuffix(f.Name(), ".conf") {
			err = json.Unmarshal(file, &data)
			if err != nil {
				logger.Get().Critical("Error unmarshalling config. error: %v", err)
				continue
			}
			collection = append(collection, data)
			SystemConfig.Provisioners[data.Provider.Name] = data.Provisioner
			data = ProviderInfo{}
		} else if strings.HasSuffix(f.Name(), ".dat") {
			if SystemConfig.SysCapabilities.ProductName != "" {
				provider_name := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
				version := make(map[string]string)
				err = json.Unmarshal(file, &version)
				if err != nil {
					logger.Get().Critical("Error unmarshalling config. error: %v", err)
					continue
				}
				SystemConfig.SysCapabilities.StorageProviderDetails[provider_name] = version["version"]
			}
		}
	}
	logger.Get().Debug("Collection: %v", collection)
	return collection
}
