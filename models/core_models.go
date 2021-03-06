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
package models

import (
	"github.com/skyrings/skyring-common/monitoring"
	"github.com/skyrings/skyring-common/tools/uuid"
	"time"
)

type Node struct {
	NodeId            uuid.UUID              `json:"nodeid"`
	Roles             []string               `json:"roles"`
	Hostname          string                 `json:"hostname"`
	Tags              []string               `json:"tags"`
	ManagementIP4     string                 `json:"management_ip4"`
	ClusterIP4        string                 `json:"cluster_ip4"`
	PublicIP4         string                 `json:"public_ip4"`
	ClusterId         uuid.UUID              `json:"clusterid"`
	Location          string                 `json:"location"`
	Status            NodeStatus             `json:"status"`
	State             NodeState              `json:"state"`
	AlmStatus         AlarmStatus            `json:"almstatus"`
	AlmWarnCount      int                    `json:"almwarncount"`
	AlmCritCount      int                    `json:"almcritcount"`
	Options           map[string]string      `json:"options"`
	CPUs              []Cpu                  `json:"cpus"`
	NetworkInfo       Network                `json:"network_info"`
	StorageDisks      []Disk                 `json:"storage_disks"`
	Memory            Memory                 `json:"memory"`
	OS                OperatingSystem        `json:"os"`
	Enabled           bool                   `json:"enabled"`
	Fingerprint       string                 `json:"saltfingerprint"`
	Utilizations      map[string]Utilization `json:"utilizations"`
	ServiceStatusList map[string][]string    `json:"servicestatuslist"`
}

type Network struct {
	IPv4   []string `bson:"ipv4",json:"ipv4"`     // TODO: use ipv4 type
	IPv6   []string `bson:"ipv6",json:"ipv6"`     // TODO: use ipv6 type
	Subnet []string `bson:"subnet",json:"subnet"` // TODO: use subnet type
}

type Disk struct {
	DevName        string    `bson:"devname",json:"devname"`
	FSType         string    `bson:"fstype",json:"fstype"`
	FSUUID         uuid.UUID `bson:"fsuuid",json:"fsuuid"`
	Model          string    `bson:"model",json:"model"`
	MountPoint     []string  `bson:"mountpoint",json:"mountpoint"`
	Name           string    `bson:"name",json:"name"`
	Parent         string    `bson:"parent",json:"parent"`
	Size           float64   `bson:"size",json:"size"`
	Type           string    `bson:"type",json:"type"`
	Used           bool      `bson:"used",json:"used"`
	SSD            bool      `bson:"ssd",json:"ssd"`
	Vendor         string    `bson:"vendor",json:"vendor"`
	StorageProfile string    `bson:"storageprofile",json:"storageprofile"`
	DiskId         uuid.UUID `bson:"diskid",json:"diskid"`
}

type Cpu struct {
	Architecture   string `bson:"architecture",json:"architecture"`
	CpuOpMode      string `bson:"cpuopmode",json:"cpuopmode"`
	CPUs           string `bson:"cpus",json:"cpus"`
	VendorId       string `bson:"vendorid",json:"vendorid"`
	ModelName      string `bson:"modelname",json:"modelname"`
	CPUFamily      string `bson:"cpufamily",json:"cpufamily"`
	CPUMHz         string `bson:"cpumhz",json:"cpumhz"`
	Model          string `bson:"model",json:"model"`
	CoresPerSocket string `bson:"corespersocket",json:"corespersocket"`
}

type OperatingSystem struct {
	Name          string `bson:"name",json:"name"`
	OSVersion     string `bson:"osversion",json:"osversion"`
	KernelVersion string `bson:"kernelversion",json:"kernelversion"`
	SELinuxMode   string `bson:"selinuxmode",json:"selinuxmode"`
}

type Memory struct {
	TotalSize string `bson:"totalsize",json:"totalsize"`
	SwapTotal string `bson:"swaptotal",json:"swaptotal"`
	Active    string `bson:"active",json:"active"`
	Type      string `bson:"type",json:"type"`
}

type User struct {
	Username            string   `json:"username"`
	Email               string   `json:"email"`
	Hash                []byte   `json:"hash"`
	Role                string   `json:"role"`
	Groups              []string `json:"groups"`
	Type                int      `json:"type"`
	Status              bool     `json:"status"`
	FirstName           string   `json:"firstname"`
	LastName            string   `json:"lastname"`
	NotificationEnabled bool     `json:"notificationenabled"`
}

type Cluster struct {
	ClusterId           uuid.UUID              `json:"clusterid"`
	Name                string                 `json:"name"`
	CompatVersion       string                 `json:"compat_version"`
	Type                string                 `json:"type"`
	WorkLoad            string                 `json:"workload"`
	Status              ClusterStatus          `json:"status"`
	Tags                []string               `json:"tags"`
	Options             map[string]interface{} `json:"options"`
	OpenStackServices   []string               `json:"openstack_services"`
	Networks            ClusterNetworks        `json:"networks"`
	Monitoring          MonitoringState        `json:"monitoring"`
	MonitoringInterval  int                    `json:"monitoringinterval"`
	State               ClusterState           `json:"state"`
	AlmStatus           AlarmStatus            `json:"almstatus"`
	AlmWarnCount        int                    `json:"almwarncount"`
	AlmCritCount        int                    `json:"almcritcount"`
	Usage               Utilization            `json:"usage"`
	StorageProfileUsage map[string]Utilization `json:"storageprofileusage"`
	ObjectCount         map[string]int64       `json:"objectcount"`
	AutoExpand          bool                   `json:"autoexpand"`
	JournalSize         string                 `json:"journalsize"`
	Utilizations        map[string]interface{} `json:"utilizations"`
}

type System struct {
	Name                      string                            `json:"name"`
	Usage                     Utilization                       `json:"usage"`
	StorageProfileUsage       map[string]map[string]interface{} `json:"storageprofileusage"`
	StorageCount              map[string]int                    `json:"storagecount"`
	SLUCount                  map[string]int                    `json:"slucount"`
	NodesCount                map[string]int                    `json:"nodescount"`
	ClustersCount             map[string]int                    `json:"clusterscount"`
	MonitoringPlugins         map[string]monitoring.Plugin      `json:"monitoringplugins"`
	ProviderMonitoringDetails map[string]map[string]interface{} `json:"providermonitoringdetails"`
	MostUsedStorages          []StorageUsage                    `json:"storageusage"`
	Utilizations              map[string]interface{}            `json:"utilizations"`
	UpdatedAt                 string                            `json:"updatedat"`
}

type Utilization struct {
	Used        int64   `json:"used"`
	Total       int64   `json:"total"`
	PercentUsed float64 `json:"percentused"`
	UpdatedAt   string  `json:"updatedat"`
}

type MonitoringState struct {
	Plugins    []monitoring.Plugin `json:"plugins"`
	StaleNodes []string            `json:"stalenodes"`
}

type ClusterNetworks struct {
	Cluster string `json:"cluster"`
	Public  string `json:"public"`
}

type StorageLogicalUnit struct {
	SluId             uuid.UUID              `json:"sluid"`
	Name              string                 `json:"name"`
	Type              int                    `json:"type"`
	ClusterId         uuid.UUID              `json:"clusterid"`
	NodeId            uuid.UUID              `json:"nodeid"`
	StorageIds        []uuid.UUID            `json:"storageid"`
	StorageDeviceId   uuid.UUID              `json:"storagedeviceid"`
	StorageDeviceSize float64                `json:"storagedevicesize"`
	Status            SluStatus              `json:"status"`
	Options           map[string]interface{} `json:"options"`
	StorageProfile    string                 `json:"storageprofile"`
	State             string                 `json:"state"`
	AlmStatus         AlarmStatus            `json:"almstatus"`
	AlmWarnCount      int                    `json:"almwarncount"`
	AlmCritCount      int                    `json:"almcritcount"`
	Usage             Utilization            `json:"usage"`
}

type Storage struct {
	StorageId           uuid.UUID         `json:"storageid"`
	Name                string            `json:"name"`
	Type                string            `json:"type"`
	Tags                []string          `json:"tags"`
	ClusterId           uuid.UUID         `json:"clusterid"`
	Size                string            `json:"size"`
	Status              StorageStatus     `json:"status"`
	Replicas            int               `json:"replicas"`
	Profile             string            `json:"profile"`
	SnapshotsEnabled    bool              `json:"snapshots_enabled"`
	SnapshotScheduleIds []uuid.UUID       `json:"snapshot_schedule_ids"`
	QuotaEnabled        bool              `json:"quota_enabled"`
	QuotaParams         map[string]string `json:"quota_params"`
	Options             map[string]string `json:"options"`
	Usage               Utilization       `json:"usage"`
	State               string            `json:"state"`
	AlmStatus           AlarmStatus       `json:"almstatus"`
	AlmWarnCount        int               `json:"almwarncount"`
	AlmCritCount        int               `json:"almcritcount"`
	SluIds              []uuid.UUID       `json:"slus"`
}

type BlockDevice struct {
	Id                  uuid.UUID         `json:"id"`
	Name                string            `json:"name"`
	Tags                []string          `json:"tags"`
	ClusterId           uuid.UUID         `json:"clusterid"`
	ClusterName         string            `json:"clustername"`
	StorageId           uuid.UUID         `json:"storageid"`
	StorageName         string            `json:"storagename"`
	Size                string            `json:"size"`
	SnapshotsEnabled    bool              `json:"snapshots_enabled"`
	SnapshotScheduleIds []uuid.UUID       `json:"snapshot_schedule_ids"`
	QuotaEnabled        bool              `json:"quota_enabled"`
	QuotaParams         map[string]string `json:"quota_params"`
	Options             map[string]string `json:"options"`
	Usage               Utilization       `json:"usage"`
	AlmStatus           AlarmStatus       `json:"almstatus"`
	AlmWarnCount        int               `json:"almwarncount"`
	AlmCritCount        int               `json:"almcritcount"`
}

type SnapshotSchedule struct {
	Id            uuid.UUID `json:"id"`
	Recurrence    string    `json:"recurrence"`
	Interval      int       `json:"interval"`
	ExecutionTime string    `json:"execution_time"`
	Days          []string  `json:"days"`
	StartFrom     string    `json:"start_from"`
	EndBy         string    `json:"endby"`
}

type Status struct {
	Timestamp time.Time
	Message   string
}

type AppTask struct {
	Id          uuid.UUID         `json:"id"`
	Name        string            `json:"name"`
	Owner       string            `json:"owner"`
	ParentId    uuid.UUID         `json:"parentid"`
	Started     bool              `json:"started"`
	Completed   bool              `json:"completed"`
	StatusList  []Status          `json:"statuslist"`
	Tag         map[string]string `json:"tag"`
	LastUpdated time.Time         `json:"lastupdated"`
	SubTasks    []uuid.UUID       `json:"subtasks"`
	Status      TaskStatus        `json:"status"`
}

type DiskProfile struct {
	Type  DiskType `json:"disktype"`
	Speed int      `json:"speed"`
}

type StorageProfile struct {
	Name     string      `json:"name"`
	Rule     DiskProfile `json:"rule"`
	Priority int         `json:"priority"`
	Default  bool        `json:"default"`
}

type ExternalUsers struct {
	Users      []User
	TotalCount int
	StartIndex int
	EndIndex   int
}

type Directory struct {
	LdapServer  string
	Type        string
	Port        uint
	Base        string
	DomainAdmin string
	Password    string
	Uid         string
	FirstName   string
	LastName    string
	DisplayName string
	Email       string
}

type ClusterNotificationSubscription struct {
	ClusterId     uuid.UUID                  `json:"clusterid"`
	Notifications []NotificationSubscription `json:"notifications"`
}
