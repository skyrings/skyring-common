/*Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the  specific language governing permissions and
limitations under the License.
*/
package models

import (
	"fmt"
	"github.com/skyrings/skyring-common/monitoring"
	"github.com/skyrings/skyring-common/tools/uuid"
	"time"
)

type AddStorageNodeRequest struct {
	Hostname       string `json:"hostname"`
	SshFingerprint string `json:"sshfingerprint"`
	User           string `json:"user"`
	Password       string `json:"password"`
	SshPort        int    `json:"sshport"`
}

type AddClusterRequest struct {
	Name               string                 `json:"name"`
	CompatVersion      string                 `json:"compat_version"`
	Type               string                 `json:"type"`
	WorkLoad           string                 `json:"workload"`
	Tags               []string               `json:"tags"`
	Options            map[string]interface{} `json:"options"`
	OpenStackServices  []string               `json:"openstackservices"`
	Nodes              []ClusterNode          `json:"nodes"`
	Networks           ClusterNetworks        `json:"networks"`
	MonitoringPlugins  []monitoring.Plugin    `json:"monitoringplugins"`
	MonitoringInterval int                    `json:"monitoringinterval"`
	DisableAutoExpand  bool                   `json:"disableautoexpand"`
	JournalSize        string                 `json:"journalSize"`
}

type ImportClusterRequest struct {
	BootstrapNode string   `json:"bootstrapnode"`
	ClusterType   string   `json:"type"`
	Nodes         []string `json:"nodes"`
}

type ClusterSummary struct {
	ClusterId                 uuid.UUID                         `json:"clusterid"`
	Name                      string                            `json:"name"`
	MostUsedStorages          []StorageUsage                    `json:"storageusage"`
	Usage                     Utilization                       `json:"usage"`
	StorageProfileUsage       map[string]map[string]interface{} `json:"storageprofileusage"`
	ObjectCount               map[string]int64                  `json:"objectcount"`
	StorageCount              map[string]int                    `json:"storagecount"`
	SLUCount                  map[string]int                    `json:"slucount"`
	MonitoringPlugins         []monitoring.Plugin               `json:"monitoringplugins"`
	NodesCount                map[string]int                    `json:"nodescount"`
	ProviderMonitoringDetails map[string]map[string]interface{} `json:"providermonitoringdetails"`
	Utilizations              map[string]interface{}            `json:"utilizations"`
}

type StorageUsage struct {
	Name  string      `json:"name"`
	Usage Utilization `json:"usage"`
}

type DiskHierarchyRequest struct {
	ClusterName  string        `json:"clustername"`
	ClusterType  string        `json:"clustertype"`
	JournalSize  string        `json:"journals"`
	ClusterNodes []ClusterNode `json:"clusternodes"`
}

type DiskHierarchyDetails struct {
	ClusterName string                       `json:"clustername"`
	Hierarchy   map[string]map[string]string `json:"hierarchy"`
	StorageSize uint64                       `json:"storagesize"`
}

type ClusterNode struct {
	NodeId   string              `json:"nodeid"`
	NodeType []string            `json:"nodetype"`
	Devices  []ClusterNodeDevice `json:"disks"`
	Options  map[string]string   `json:"options"`
}

type ClusterNodeDevice struct {
	Name    string            `json:"name"`
	FSType  string            `json:"fstype"`
	Options map[string]string `json:"options"`
}

type AddStorageRequest struct {
	Name             string                         `json:"name"`
	Type             string                         `json:"type"`
	Tags             []string                       `json:"tags"`
	Size             string                         `json:"size"`
	Replicas         int                            `json:"replicas"`
	Profile          string                         `json:"profile"`
	SnapshotsEnabled bool                           `json:"snapshots_enabled"`
	SnapshotSchedule SnapshotScheduleRequest        `json:"snapshot_schedule"`
	QuotaEnabled     bool                           `json:"quota_enabled"`
	QuotaParams      map[string]string              `json:"quota_params"`
	Options          map[string]string              `json:"options"`
	BlockDevices     []AddStorageBlockDeviceRequest `json:"blockdevices"`
}

type AddStorageBlockDeviceRequest struct {
	Name             string                  `json:"name"`
	Tags             []string                `json:"tags"`
	Size             string                  `json:"size"`
	SnapshotsEnabled bool                    `json:"snapshots_enabled"`
	SnapshotSchedule SnapshotScheduleRequest `json:"snapshot_schedule"`
	QuotaEnabled     bool                    `json:"quota_enabled"`
	QuotaParams      map[string]string       `json:"quota_params"`
	Options          map[string]string       `json:"options"`
}

type SnapshotScheduleRequest struct {
	Recurrence    string   `json:"recurrence"`
	Interval      int      `json:"interval"`
	ExecutionTime string   `json:"execution_time"`
	Days          []string `json:"days"`
	StartFrom     string   `json:"start_from"`
	EndBy         string   `json:"endby"`
}

type Nodes []Node

type NodeEvent struct {
	Timestamp time.Time         `json:"timestamp"`
	Node      string            `json:"node"`
	Tag       string            `json:"tag"`
	Tags      map[string]string `json:"tags"`
	Message   string            `json:"message"`
	Severity  string            `json:"severity"`
}

type Event struct {
	EventId         uuid.UUID          `json:"event_id"`
	ClusterId       uuid.UUID          `json:"cluster_id"`
	NodeId          uuid.UUID          `json:"node_id"`
	Timestamp       time.Time          `json:"timestamp"`
	Tag             string             `json:"tag"`
	Tags            map[string]string  `json:"tags"`
	Message         string             `json:"message"`
	Severity        string             `json:"severity"`
	ImpactingEvents map[string][]Event `json:"impactingentities"`
}

type ThresholdEvent struct {
	ClusterId         uuid.UUID `json:"clusterid"`
	UtilizationType   string    `json:"utilizationtype"`
	ThresholdSeverity string    `json:"thresholdseverity"`
	TimeStamp         time.Time `json:"timestamp"`
	EntityId          uuid.UUID `json:"entityid"`
	EntityName        string    `json:"entityname"`
}

type AppEvent struct {
	EventId            uuid.UUID          `json:"event_id"`
	ClusterId          uuid.UUID          `json:"cluster_id"`
	ClusterName        string             `json:"cluster_name"`
	NotificationEntity NotificationEntity `json:"notificationentity"`
	EntityId           uuid.UUID          `json:"entityid"`
	NodeId             uuid.UUID          `json:"node_id"`
	NodeName           string             `json:"nodename"`
	Timestamp          time.Time          `json:"timestamp"`
	Name               string             `json:"name"`
	Context            string             `json:"context"`
	Tags               map[string]string  `json:"tags"`
	Message            string             `json:"message"`
	Description        string             `json:"description"`
	Severity           AlarmStatus        `json:"severity"`
	CorrelationId      uuid.UUID          `json:"correlationid"`
	Acked              bool               `json:"acked"`
	UserAckedTime      time.Time          `json:"userackedtime"`
	SystemAckedTime    time.Time          `json:"systemackedtime"`
	AckedByUser        string             `json:"ackedbyuser"`
	AckedByEvent       string             `json:"ackedbyevent"`
	UserAckComment     string             `json:"userackcomment"`
	SystemAckComment   string             `json:"systemackcomment"`
	Notify             bool               `json:"notify"`
	Notified           bool               `json:"notified"`
}

type MailNotifier struct {
	MailId           string `json:"mailid"`
	Passcode         string `json:"passcode"`
	SmtpServer       string `json:"smtpserver"`
	Port             int    `json:"port"`
	Encryption       string `json:"encryption"`
	SkipVerify       bool   `json:"skipverify"`
	MailNotification bool   `json:"mailnotification"`
	SubPrefix        string `json:"subprefix"`
}

type QueryOps struct {
	Sort     bool
	Batch    int
	Iter     bool
	Limit    int
	Prefetch float64
	Select   interface{}
	Skip     bool
	Distinct bool
}

type ApiRoute struct {
	Name    string
	Pattern string
	Method  string
	Version int
}

type ClusterForImport struct {
	ClusterName string          `json:"clustername"`
	ClusterId   uuid.UUID       `json:"clusterid"`
	Version     string          `json:"version"`
	Compatible  bool            `json:"compatible"`
	Nodes       []NodeForImport `json:"nodes"`
}

type NodeForImport struct {
	Name  string   `json:"name"`
	Type  []string `json:"type"`
	Found bool     `json:"found"`
}

const (
	DEFAULT_SSH_PORT   = 22
	DEFAULT_FS_TYPE    = "xfs"
	ENGINE_NAME        = "skyring"
	NODE_TYPE_OSD      = "OSD"
	REQUEST_SIZE_LIMIT = 1048576

	COLL_NAME_STORAGE                            = "storage"
	COLL_NAME_NODE_EVENTS                        = "node_events"
	COLL_NAME_APP_EVENTS                         = "app_events"
	COLL_NAME_STORAGE_NODES                      = "storage_nodes"
	COLL_NAME_STORAGE_CLUSTERS                   = "storage_clusters"
	COLL_NAME_SKYRING_UTILIZATION                = "skyring_utilization"
	COLL_NAME_CLUSTER_SUMMARY                    = "cluster_summary"
	COLL_NAME_STORAGE_LOGICAL_UNITS              = "storage_logical_units"
	COLL_NAME_TASKS                              = "tasks"
	COLL_NAME_SESSION_STORE                      = "skyring_session_store"
	COLL_NAME_USER                               = "skyringusers"
	COLL_NAME_STORAGE_PROFILE                    = "storage_profile"
	COLL_NAME_MAIL_NOTIFIER                      = "mailnotifier"
	COLL_NAME_LDAP                               = "ldap"
	COLL_NAME_BLOCK_DEVICES                      = "block_devices"
	COLL_NAME_SYSTEM_CAPABILITIES                = "system_capabilities"
	COLL_NAME_THRESHOLD_BREACHES                 = "threshold_breaches"
	COLL_NAME_CLUSTER_NOTIFICATION_SUBSCRIPTIONS = "cluster_notification_subscriptions"
	COLL_NAME_ARCHIVE_TASKS                      = "archive_tasks"
	COLL_NAME_ARCHIVE_EVENTS                     = "archive_events"

	TASKS_PER_PAGE      = 100
	LDAP_USERS_PER_PAGE = 100
	EVENTS_PER_PAGE     = 100

	STORAGE_TYPE_REPLICATED    = "replicated"
	STORAGE_TYPE_ERASURE_CODED = "erasure_coded"
	Monitor                    = "monitor"
	Mon                        = "mon"
	Yes                        = "Y"
	No                         = "N"
	TotalSLU                   = "TotalSLU"
	ErrorSLU                   = "ErrorSLU"
	WarningSLU                 = "WarningSLU"
	DownSLU                    = "DownSLU"
	SLU_STATE_DOWN             = "Out"
)

type Clusters []Cluster
type Storages []Storage

type UnmanagedNode struct {
	Name            string `json:"name"`
	SaltFingerprint string `json:"saltfingerprint"`
}

type UnmanagedNodes []UnmanagedNode

type ClusterStatus int

// Status values for the cluster
const (
	CLUSTER_STATUS_OK = iota
	CLUSTER_STATUS_WARN
	CLUSTER_STATUS_ERROR
	CLUSTER_STATUS_UNKNOWN
)

var ClusterStatuses = [...]string{
	"ok",
	"warning",
	"error",
	"unknown",
}

type ClusterState int

// State values for cluster
const (
	CLUSTER_STATE_CREATING = iota
	CLUSTER_STATE_FAILED
	CLUSTER_STATE_ACTIVE
	CLUSTER_STATE_UNMANAGED
	CLUSTER_STATE_SYNCING
)

var ClusterStates = [...]string{
	"creating",
	"failed",
	"active",
	"unmanaged",
	"syncing",
}

func (s ClusterState) String() string { return ClusterStates[s] }

// Storage logical unit types
const (
	CEPH_OSD = 1 + iota
)

var StorageLogicalUnitTypes = [...]string{
	"osd",
}

const (
	STATUS_UP   = "up"
	STATUS_DOWN = "down"
	STATUS_OK   = "ok"
	STATUS_WARN = "warning"
	STATUS_ERR  = "error"
)

func (c ClusterStatus) String() string { return ClusterStatuses[c-1] }

type AsyncResponse struct {
	TaskId uuid.UUID `json:"taskid"`
}

func (s Status) String() string {
	return fmt.Sprintf("%s %s", s.Timestamp, s.Message)
}

type TaskStatus int

const (
	TASK_STATUS_NONE = iota
	TASK_STATUS_SUCCESS
	TASK_STATUS_TIMED_OUT
	TASK_STATUS_FAILURE
)

var TaskStatuses = [...]string{
	"none",
	"success",
	"timedout",
	"failed",
}

func (t TaskStatus) String() string { return TaskStatuses[t] }

type DiskType int

const (
	NONE = iota
	SAS
	SSD
)

var DiskTypes = [...]string{
	"none",
	"sas",
	"ssd",
}

func (d DiskType) String() string { return DiskTypes[d] }

const (
	DefaultProfile1 = "sas"
	DefaultProfile2 = "ssd"
	DefaultProfile3 = "general"
	DefaultPriority = 100
)

type NodeState int

const (
	NODE_STATE_UNACCEPTED = iota
	NODE_STATE_INITIALIZING
	NODE_STATE_ACTIVE
	NODE_STATE_FAILED
	NODE_STATE_UNMANAGED
	NODE_STATE_IMPORTING
)

var NodeStates = [...]string{
	"unaccepted",
	"initializing",
	"active",
	"failed",
	"unmanaged",
	"importing",
}

const (
	NODES   = "nodes"
	CLUSTER = "cluster"
	TOTAL   = "total"
	SYSTEM  = "system"
)

func (s NodeState) String() string { return NodeStates[s] }

type NodeStatus int

// Status values for the cluster
const (
	NODE_STATUS_OK = iota
	NODE_STATUS_WARN
	NODE_STATUS_ERROR
	NODE_STATUS_UNKNOWN
)

var NodeStatuses = [...]string{
	"ok",
	"warning",
	"error",
	"unknown",
}

func (s NodeStatus) String() string { return NodeStatuses[s] }

type AlarmStatus int

// Status values for the cluster
const (
	ALARM_STATUS_INDETERMINATE = iota
	ALARM_STATUS_CRITICAL
	ALARM_STATUS_MAJOR
	ALARM_STATUS_MINOR
	ALARM_STATUS_WARNING
	ALARM_STATUS_CLEARED
)

var AlarmStatuses = [...]string{
	"indeterminate",
	"critical",
	"major",
	"minor",
	"warning",
	"cleared",
}

func (s AlarmStatus) String() string { return AlarmStatuses[s] }

type NotificationEntity int

// types of notification entities
const (
	NOTIFICATION_ENTITY_HOST = iota
	NOTIFICATION_ENTITY_CLUSTER
	NOTIFICATION_ENTITY_SLU
	NOTIFICATION_ENTITY_STORAGE
	NOTIFICATION_ENTITY_USER
	NOTIFICATION_ENTITY_BLOCK_DEVICE
	NOTIFICATION_ENTITY_STORAGE_PROFILE
	NOTIFICATION_ENTITY_MAIL_NOTIFIER
)

var NotificationEntities = [...]string{
	"Host",
	"Cluster",
	"Slu",
	"Storage",
	"User",
	"Block Device",
	"Storage Profile",
	"Mail Notifier",
}

func (s NotificationEntity) String() string { return NotificationEntities[s] }

type SluStatus int

// Status values for the cluster
const (
	SLU_STATUS_OK = iota
	SLU_STATUS_WARN
	SLU_STATUS_ERROR
	SLU_STATUS_UNKNOWN
)

var SluStatuses = [...]string{
	"ok",
	"warning",
	"error",
	"unknown",
}

func (s SluStatus) String() string { return SluStatuses[s] }

type StorageStatus int

// Status values for the cluster
const (
	STORAGE_STATUS_OK = iota
	STORAGE_STATUS_WARN
	STORAGE_STATUS_ERROR
	STORAGE_STATUS_UNKNOWN
)

var StorageStatuses = [...]string{
	"ok",
	"warning",
	"error",
	"unknown",
}

const (
	CURRENT_VALUE     = "CurrentValue"
	CLUSTER_ID        = "ClusterId"
	ENTITY_ID         = "EntityId"
	PLUGIN            = "Plugin"
	THRESHOLD_TYPE    = "ThresholdType"
	THRESHOLD_VALUE   = "ThresholdValue"
	ENTITY_NAME       = "EntityName"
	CLUSTER_CONFIGS   = "cluster_configs"
	THRESHOLD_CONFIGS = "threshold_configs"
	NOTIFICATION_LIST = "notification_list"
	UTILIZATIONS      = "utilizations"
	NOTIFY            = "Notify"
	WARNING           = "WARNING"
	OK                = "OK"
	CRITICAL          = "CRITICAL"
	NEAR_FULL         = "nearfull"
)

func (s StorageStatus) String() string { return StorageStatuses[s] }

type CollectdSingleValuedMetric struct {
	Used        string `json:"Used"`
	Total       string `json:"Total"`
	PercentUsed string `json:"PercentUsed"`
}

var SkyringServices = [...]string{
	"collectd",
	"salt-minion",
}

type CollectdCpuMetric struct {
	PercentUsed string `json:"PercentUsed"`
}
