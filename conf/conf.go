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
	"github.com/op/go-logging"
	"github.com/skyrings/skyring-common/tools/logger"
	"io/ioutil"
	"path"
)

const (
	aboutConfigFile = "about.conf"
)

type SkyringConfig struct {
	Host              string `json:"host"`
	HttpPort          int    `json:"httpPort"`
	SupportedVersions []int  `json:"supportedversions"`
}

type SkyringLogging struct {
	LogToStderr bool          `json:"logToStderr"`
	Filename    string        `json:"filename"`
	Level       logging.Level `json:"level"`
}

type AuthConfig struct {
	ProviderName string
	ConfigFile   string
}

type ProvisionerConfig struct {
	ProvisionerName string `json:"provisionername"`
	ConfigFilePath  string `json:"configfilepath"`
	RedhatStorage   bool   `json:"redhatstorage"`
	RedhatUseCdn    bool   `json:"redhatusecdn"`
}

type SkyringCollection struct {
	Config               SkyringConfig                `json:"config"`
	Logging              SkyringLogging               `json:"logging"`
	NodeManagementConfig NodeManagerConfig            `json:"nodemanagementconfig"`
	DBConfig             AppDBConfig                  `json:"dbconfig"`
	TimeSeriesDBConfig   MonitoringDBconfig           `json:"timeseriesdbconfig"`
	Authentication       AuthConfig                   `json:"authentication"`
	SummaryConfig        SystemSummaryConfig          `json:"summary"`
	Provisioners         map[string]ProvisionerConfig `json:"provisioners"`
	SysCapabilities      SystemCapabilities           `json:"systemcapabilities"`
}

type SystemSummaryConfig struct {
	NetSummaryInterval int `json:"netSummaryInterval"`
}

type AppDBConfig struct {
	Hostname string `json:"hostname"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type MonitoringDBconfig struct {
	Hostname       string `json:"hostname"`
	Port           int    `json:"port"`
	DataPushPort   int    `json:"dataPushPort"`
	CollectionName string `json:"collection_name"`
	User           string `json:"user"`
	Password       string `json:"password"`
	ManagerName    string `json:"managername"`
	ConfigFilePath string `json:"configfilepath"`
}

type NodeManagerConfig struct {
	ManagerName    string `json:"managername"`
	ConfigFilePath string `json:"configfilepath"`
}

type SystemCapabilities struct {
	ProductName            string            `bson:"productname",json:"productname"`
	ProductVersion         string            `bson:"productversion",json:"productversion"`
	StorageProviderDetails map[string]string `bson:"providerversion",json:"providerversion"`
	DbSoftware             string            `bson:"dbsoftware",json:"dbsoftware"`
	DbSoftwareVersion      string            `bson:"dbsoftwareversion",json:"dbsoftwareversion"`
}

var (
	SystemConfig SkyringCollection
)

func LoadAppConfiguration(configFilePath string, configFile string) {
	file, err := ioutil.ReadFile(path.Join(configFilePath, configFile))
	if err != nil {
		logger.Get().Critical("Error reading skyring config. error: %v", err)
		return
	}
	err = json.Unmarshal(file, &SystemConfig)
	if err != nil {
		logger.Get().Critical("Error unmarshalling skyring config. error: %v", err)
		return
	}
	//Initialize the Provisioner Map
	SystemConfig.Provisioners = make(map[string]ProvisionerConfig)
	//Initialize System_capabilities
	file, err = ioutil.ReadFile(path.Join(configFilePath, aboutConfigFile))
	if err != nil {
		logger.Get().Critical("Error reading about config. error: %v", err)
		return
	}
	err = json.Unmarshal(file, &SystemConfig.SysCapabilities)
	if err != nil {
		logger.Get().Critical("Error unmarshalling about config. error: %v", err)
		return
	}
	SystemConfig.SysCapabilities.StorageProviderDetails = make(map[string]string)
}
