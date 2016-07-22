package util

import (
	"fmt"
	"strconv"
	"time"

	"github.com/skyrings/skyring-common/conf"
	"github.com/skyrings/skyring-common/db"
	"github.com/skyrings/skyring-common/models"
	"github.com/skyrings/skyring-common/monitoring"
	"github.com/skyrings/skyring-common/tools/logger"
	"github.com/skyrings/skyring-common/tools/uuid"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func GetClusters(selectCriteria bson.M) ([]models.Cluster, error) {
	var clusters []models.Cluster

	sessionCopy := db.GetDatastore().Copy()
	defer sessionCopy.Close()

	collection := sessionCopy.DB(conf.SystemConfig.DBConfig.Database).C(models.COLL_NAME_STORAGE_CLUSTERS)
	err := collection.Find(selectCriteria).All(&clusters)
	return clusters, err
}

func GetClusterSummaries(selectCriteria bson.M) ([]models.ClusterSummary, error) {
	var clustersSummary []models.ClusterSummary

	sessionCopy := db.GetDatastore().Copy()
	defer sessionCopy.Close()

	collection := sessionCopy.DB(conf.SystemConfig.DBConfig.Database).C(models.COLL_NAME_CLUSTER_SUMMARY)
	err := collection.Find(selectCriteria).All(&clustersSummary)
	return clustersSummary, err
}

func GetSystem() (system models.System, err error) {
	sessionCopy := db.GetDatastore().Copy()
	defer sessionCopy.Close()

	collection := sessionCopy.DB(conf.SystemConfig.DBConfig.Database).C(models.COLL_NAME_SKYRING_UTILIZATION)
	err = collection.Find(nil).One(&system)
	return system, err
}

func GetTopStorageUsage(selectCriteria bson.M) ([]models.StorageUsage, error) {
	sessionCopy := db.GetDatastore().Copy()
	defer sessionCopy.Close()

	coll := sessionCopy.DB(conf.SystemConfig.DBConfig.Database).C(models.COLL_NAME_STORAGE)
	var stotrageUsage []models.StorageUsage
	var mostUsedStorages []models.StorageUsage
	if err := coll.Find(selectCriteria).Sort("-percentused").All(&stotrageUsage); err != nil {
		return mostUsedStorages, err
	}
	if len(stotrageUsage) > 5 {
		mostUsedStorages = stotrageUsage[:4]
	} else {
		mostUsedStorages = stotrageUsage
	}
	if len(stotrageUsage) == 0 {
		mostUsedStorages = make([]models.StorageUsage, 0)
	}
	return mostUsedStorages, nil
}

func GetStorageCount(selectCriteria bson.M) (map[string]int, error) {
	sessionCopy := db.GetDatastore().Copy()
	defer sessionCopy.Close()

	coll := sessionCopy.DB(conf.SystemConfig.DBConfig.Database).C(models.COLL_NAME_STORAGE)
	var storages models.Storages
	err := coll.Find(selectCriteria).All(&storages)
	storage_down_cnt := 0
	storageCriticalAlertsCount := 0
	for _, storage := range storages {
		storageCriticalAlertsCount = storageCriticalAlertsCount + storage.AlmCritCount
		if storage.Status == models.STORAGE_STATUS_ERROR {
			storage_down_cnt = storage_down_cnt + 1
		}
	}
	return map[string]int{models.TOTAL: len(storages), models.STATUS_DOWN: storage_down_cnt, models.CriticalAlerts: storageCriticalAlertsCount}, err
}

func ComputeUsage(selectCriteria bson.M) (models.Utilization, error) {
	var used int64
	var total int64
	var percentUsed float64

	clusters, clustersFetchErr := GetClusters(selectCriteria)
	if clustersFetchErr != nil {
		return models.Utilization{}, clustersFetchErr
	} else {
		for _, cluster := range clusters {
			used = used + cluster.Usage.Used
			total = total + cluster.Usage.Total
		}
		if total > 0 {
			percentUsed = float64((used * 100)) / float64(total)
		}
		return models.Utilization{Used: used, Total: total, PercentUsed: percentUsed, UpdatedAt: time.Now().String()}, nil
	}
}

func ComputeSluStatusWiseCount(sluSelectCriteria bson.M, sluThresholdSelectCriteria bson.M) (map[string]int, error) {
	var err_str string
	sessionCopy := db.GetDatastore().Copy()
	defer sessionCopy.Close()
	coll := sessionCopy.DB(conf.SystemConfig.DBConfig.Database).C(models.COLL_NAME_STORAGE_LOGICAL_UNITS)
	var slus []models.StorageLogicalUnit
	err := coll.Find(sluSelectCriteria).All(&slus)
	if err != nil && err != mgo.ErrNotFound {
		err_str = fmt.Sprintf("%s", err.Error())
	}
	slu_down_cnt := 0
	slu_error_cnt := 0
	sluThresholdEventsInDb, err := fetchThresholdEvents(sluThresholdSelectCriteria, monitoring.SLU_UTILIZATION)
	if err != nil && err != mgo.ErrNotFound {
		err_str = fmt.Sprintf("%s.%s", err_str, err.Error())
	}

	sluCriticalAlertCount := 0
	for _, slu := range slus {
		sluCriticalAlertCount = sluCriticalAlertCount + slu.AlmCritCount
		if slu.Status == models.SLU_STATUS_ERROR {
			slu_error_cnt = slu_error_cnt + 1
		}
		if slu.State == models.SLU_STATE_DOWN {
			slu_down_cnt = slu_down_cnt + 1
		}
	}
	if err_str == "" {
		err = nil
	} else {
		err = fmt.Errorf("%s", err_str)
	}
	return map[string]int{models.TOTAL: len(slus), models.SluStatuses[models.SLU_STATUS_ERROR]: slu_error_cnt, models.STATUS_DOWN: slu_down_cnt, models.NEAR_FULL: len(sluThresholdEventsInDb), models.CriticalAlerts: sluCriticalAlertCount}, err
}

func ComputeClustersStatusWiseCounts() (map[string]int, error) {
	var err_str string
	var clusters_in_error, clusters_in_warn int
	clusterCriticalAlertCount := 0
	nearFullClusters := 0
	clusters, err := GetClusters(nil)
	if err != nil && err != mgo.ErrNotFound {
		err_str = fmt.Sprintf("%s", err.Error())
	}
	selectCriteria := bson.M{
		"utilizationtype":   monitoring.CLUSTER_UTILIZATION,
		"thresholdseverity": models.CRITICAL,
	}
	clusterThresholdEvenstInDb, err := fetchThresholdEvents(selectCriteria, monitoring.CLUSTER_UTILIZATION)
	if err != nil && err != mgo.ErrNotFound {
		err_str = fmt.Sprintf("%s.%s", err_str, err.Error())
	}

	for _, cluster := range clusters {
		if cluster.Status == models.CLUSTER_STATUS_ERROR {
			clusters_in_error = clusters_in_error + 1
		} else if cluster.Status == models.CLUSTER_STATUS_WARN {
			clusters_in_warn = clusters_in_warn + 1
		}
		clusterCriticalAlertCount = clusterCriticalAlertCount + cluster.AlmCritCount
		for _, tEvent := range clusterThresholdEvenstInDb {
			if uuid.Equal(tEvent.EntityId, cluster.ClusterId) {
				nearFullClusters = nearFullClusters + 1
			}
		}
	}
	if err_str == "" {
		err = nil
	} else {
		err = fmt.Errorf("%s", err_str)
	}
	return map[string]int{models.TOTAL: len(clusters), models.ClusterStatuses[models.CLUSTER_STATUS_ERROR]: clusters_in_error, models.ClusterStatuses[models.CLUSTER_STATUS_WARN]: clusters_in_warn, models.NEAR_FULL: nearFullClusters, models.CriticalAlerts: clusterCriticalAlertCount}, err
}

func fetchThresholdEvents(selectCriteria bson.M, utilizationType string) ([]models.ThresholdEvent, error) {
	var tEventsInDb []models.ThresholdEvent
	sessionCopy := db.GetDatastore().Copy()
	defer sessionCopy.Close()
	coll := sessionCopy.DB(conf.SystemConfig.DBConfig.Database).C(models.COLL_NAME_THRESHOLD_BREACHES)
	err := coll.Find(selectCriteria).All(&tEventsInDb)
	return tEventsInDb, err
}

func ComputeStorageProfileUtilization(selectCriteria bson.M, spThresholdconfigs []monitoring.PluginConfig) (map[string]map[string]interface{}, error) {
	net_storage_profile_utilization := make(map[string]map[string]interface{})
	clusters, err := GetClusters(selectCriteria)
	for _, cluster := range clusters {
		for profile, profileUtilization := range cluster.StorageProfileUsage {
			used := profileUtilization.Used
			total := profileUtilization.Total
			if utilization, ok := net_storage_profile_utilization[profile]["utilization"].(models.Utilization); ok {
				used = used + utilization.Used
				total = total + utilization.Total
			}
			var percentUsed float64
			if total != 0.0 {
				percentUsed = float64(used*100) / float64(total)
			}
			net_storage_profile_utilization[profile] = map[string]interface{}{"utilization": models.Utilization{Used: used, Total: total, PercentUsed: percentUsed, UpdatedAt: time.Now().String()}}
		}
	}
	if len(net_storage_profile_utilization) != 0 {
		var spNearFullThreshold float64
		var spFullThreshold float64
		for _, spThreshold := range spThresholdconfigs {
			if spThreshold.Type == monitoring.CRITICAL {
				spFullThreshold, _ = strconv.ParseFloat(spThreshold.Value, 64)
			}
			if spThreshold.Type == monitoring.WARNING {
				spNearFullThreshold, _ = strconv.ParseFloat(spThreshold.Value, 64)
			}
		}
		for sProfile, isNearFullUtilizationMap := range net_storage_profile_utilization {
			net_storage_profile_utilization[sProfile]["isNearFull"] = false
			net_storage_profile_utilization[sProfile]["isFull"] = false
			if isNearFullUtilizationMap["utilization"] != nil {
				if isNearFullUtilizationMap["utilization"].(models.Utilization).PercentUsed > spFullThreshold {
					net_storage_profile_utilization[sProfile]["isFull"] = true
					continue
				}
				if isNearFullUtilizationMap["utilization"].(models.Utilization).PercentUsed > spNearFullThreshold {
					net_storage_profile_utilization[sProfile]["isNearFull"] = true
				}
			}
		}
	}
	return net_storage_profile_utilization, err
}

func InitializeClusterSummary(cluster models.Cluster) {
	reqId, err := uuid.New()
	if err != nil {
		logger.Get().Error("Error Creating the RequestId. error: %v", err)
		return
	}

	ctxt := fmt.Sprintf("%v:%v", models.ENGINE_NAME, reqId.String())

	cSummary := models.ClusterSummary{}

	if mostUsedStorages, err := GetTopStorageUsage(bson.M{"clusterid": cluster.ClusterId}); err == nil {
		cSummary.MostUsedStorages = mostUsedStorages
	}

	sluStatusWiseCounts, err := ComputeSluStatusWiseCount(
		bson.M{"clusterid": cluster.ClusterId},
		bson.M{"utilizationtype": monitoring.SLU_UTILIZATION, "clusterid": cluster.ClusterId})
	if err == nil {
		cSummary.SLUCount = sluStatusWiseCounts
	}

	cSummary.Usage = cluster.Usage

	cSummary.ObjectCount = cluster.ObjectCount

	cSummary.Utilizations = cluster.Utilizations

	storageCount, err := GetStorageCount(bson.M{"clusterid": cluster.ClusterId})
	if err != nil {
		logger.Get().Error("%s - Failed to fetch storage status wise counts for cluster %v.Error %v", ctxt, cluster.Name, err)
	}
	cSummary.StorageCount = storageCount

	cSummary.ClusterId = cluster.ClusterId
	cSummary.Type = cluster.Type
	cSummary.Name = cluster.Name
	cSummary.MonitoringPlugins = cluster.Monitoring.Plugins
	cSummary.UpdatedAt = time.Now().String()

	sessionCopy := db.GetDatastore().Copy()
	defer sessionCopy.Close()

	coll := sessionCopy.DB(conf.SystemConfig.DBConfig.Database).C(models.COLL_NAME_CLUSTER_SUMMARY)
	if _, err := coll.Upsert(bson.M{"clusterid": cluster.ClusterId}, cSummary); err != nil {
		logger.Get().Error("%s - Error persisting the cluster summary.Error %v", ctxt, err)
	}
}

func UpdateDb(selectCriteria bson.M, update bson.M, collectionName string, ctxt string) {
	sessionCopy := db.GetDatastore().Copy()
	defer sessionCopy.Close()

	coll := sessionCopy.DB(conf.SystemConfig.DBConfig.Database).C(collectionName)
	if err := coll.Update(selectCriteria, bson.M{"$set": update}); err != nil {
		logger.Get().Error("%s - Failed to perform the update for select criteria %v and update %v. Error %v", ctxt, selectCriteria, update, err)
	}
}

func UpdateStorageCountToSummaries(ctxt string, cluster models.Cluster) {
	storageCnt, err := GetStorageCount(bson.M{"clusterid": cluster.ClusterId})
	if err != nil {
		logger.Get().Error("%s - Failed to fetch storage count for cluster %v. Error %v", ctxt, cluster.Name, err)
	} else {
		UpdateDb(bson.M{"clusterid": cluster.ClusterId}, bson.M{"storagecount": storageCnt}, models.COLL_NAME_CLUSTER_SUMMARY, ctxt)
		storageCnt, err := GetStorageCount(nil)
		if err != nil {
			logger.Get().Error("%s - Failed to update storage state wise count for system summary.Error %v", ctxt, err)
		} else {
			UpdateDb(bson.M{"name": monitoring.SYSTEM}, bson.M{"storagecount": storageCnt}, models.COLL_NAME_SKYRING_UTILIZATION, ctxt)
		}
	}
}

func UpdateClusterStateWiseCount(ctxt string) {
	clusterCount, err := ComputeClustersStatusWiseCounts()
	if err != nil {
		logger.Get().Error("%s - Failed to update state wise clusters count to summary", ctxt)
	} else {
		UpdateDb(bson.M{"name": monitoring.SYSTEM}, bson.M{"clusterscount": clusterCount}, models.COLL_NAME_SKYRING_UTILIZATION, ctxt)
	}
}
func UpdateSluCountToSummaries(ctxt string, cluster models.Cluster) {
	sluCnt, err := ComputeSluStatusWiseCount(bson.M{"clusterid": cluster.ClusterId}, bson.M{"utilizationtype": monitoring.SLU_UTILIZATION, "clusterid": cluster.ClusterId, "thresholdseverity": models.CRITICAL})
	if err != nil {
		logger.Get().Error("%s - Failed to fetch slu status wise count for cluster %v.Error %v", ctxt, cluster.Name, err)
	} else {
		UpdateDb(bson.M{"clusterid": cluster.ClusterId}, bson.M{"slucount": sluCnt}, models.COLL_NAME_CLUSTER_SUMMARY, ctxt)
		sluCnt, err := ComputeSluStatusWiseCount(nil, bson.M{"utilizationtype": monitoring.SLU_UTILIZATION, "thresholdseverity": models.CRITICAL})
		if err != nil {
			logger.Get().Error("%s - Failed to update slu status wise count for system summary.Error %v", ctxt, err)
		} else {
			UpdateDb(bson.M{"name": monitoring.SYSTEM}, bson.M{"slucount": sluCnt}, models.COLL_NAME_SKYRING_UTILIZATION, ctxt)
		}
	}
}

func UpdateStorageProfileUtilizationToSummaries(ctxt string, cluster models.Cluster) {
	pluginIndex := monitoring.GetPluginIndex(monitoring.STORAGE_PROFILE_UTILIZATION, cluster.Monitoring.Plugins)
	if pluginIndex != -1 {
		if storageProfileUtilization, err := ComputeStorageProfileUtilization(
			bson.M{"clusterid": cluster.ClusterId},
			cluster.Monitoring.Plugins[pluginIndex].Configs); err != nil {
			logger.Get().Error("%s - Failed to fetch storage profile utilization of cluster %v.Error %v", ctxt, cluster.Name, err)
		} else {
			UpdateDb(bson.M{"clusterid": cluster.ClusterId}, bson.M{"storageprofileusage": storageProfileUtilization}, models.COLL_NAME_CLUSTER_SUMMARY, ctxt)
		}
	}

	systemthresholds := monitoring.GetSystemDefaultThresholdValues()
	utilization, err := ComputeStorageProfileUtilization(nil, systemthresholds[monitoring.STORAGE_PROFILE_UTILIZATION].Configs)
	if err != nil {
		logger.Get().Error("%s - Error updating the storage profile utilization to system summary. Error %v", ctxt, err)
	} else {
		UpdateDb(bson.M{"name": monitoring.SYSTEM}, bson.M{"storageprofileusage": utilization}, models.COLL_NAME_SKYRING_UTILIZATION, ctxt)
	}
}

func FetchNodeStatusWiseCounts(selectCriteria bson.M) (map[string]int, error) {
	sessionCopy := db.GetDatastore().Copy()
	defer sessionCopy.Close()
	var nodes []models.Node
	error_nodes := 0
	nodeCriticalAlertCount := 0
	unmanagedNodesCount := 0
	coll := sessionCopy.DB(conf.SystemConfig.DBConfig.Database).C(models.COLL_NAME_STORAGE_NODES)
	if nodesError := coll.Find(nil).All(&nodes); nodesError != nil {
		if nodesError != mgo.ErrNotFound {
			return map[string]int{models.TOTAL: len(nodes), models.NodeStatuses[models.NODE_STATUS_ERROR]: error_nodes, models.NodeStates[models.NODE_STATE_UNACCEPTED]: unmanagedNodesCount, models.CriticalAlerts: nodeCriticalAlertCount}, fmt.Errorf("Failed to fetch nodes. error: %v", nodesError)
		}
	}
	for _, node := range nodes {
		nodeCriticalAlertCount = nodeCriticalAlertCount + node.AlmCritCount
		if node.Status == models.NODE_STATUS_ERROR {
			error_nodes = error_nodes + 1
		}
		if node.State == models.NODE_STATE_UNACCEPTED {
			unmanagedNodesCount = unmanagedNodesCount + 1
		}
	}
	return map[string]int{models.TOTAL: len(nodes), models.NodeStatuses[models.NODE_STATUS_ERROR]: error_nodes, models.NodeStates[models.NODE_STATE_UNACCEPTED]: unmanagedNodesCount, models.CriticalAlerts: nodeCriticalAlertCount}, nil
}

func InitializeSystemSummary() {
	var system models.System

	reqId, err := uuid.New()
	if err != nil {
		logger.Get().Error("Error Creating the RequestId. error: %v", err)
		return
	}

	ctxt := fmt.Sprintf("%v:%v", models.ENGINE_NAME, reqId.String())

	system.Name = monitoring.SYSTEM

	_, clusterFetchError := GetClusters(nil)
	if clusterFetchError != nil {
		if clusterFetchError == mgo.ErrNotFound {
			return
		}
		logger.Get().Error("%s - Failed to fetch clusters.Err %v", ctxt, clusterFetchError)
	}

	sluStatusWiseCounts, err := ComputeSluStatusWiseCount(nil, bson.M{"utilizationtype": monitoring.SLU_UTILIZATION, "thresholdseverity": models.CRITICAL})
	system.SLUCount = sluStatusWiseCounts

	storageCount, err := GetStorageCount(nil)
	system.StorageCount = storageCount

	clustersCount, err := ComputeClustersStatusWiseCounts()
	system.ClustersCount = clustersCount
	systemthresholds := monitoring.GetSystemDefaultThresholdValues()

	net_storage_profile_utilization, err := ComputeStorageProfileUtilization(nil, systemthresholds[monitoring.STORAGE_PROFILE_UTILIZATION].Configs)
	system.StorageProfileUsage = net_storage_profile_utilization

	system.UpdatedAt = time.Now().String()

	mostUsedStorages, err := GetTopStorageUsage(nil)
	system.MostUsedStorages = mostUsedStorages

	sessionCopy := db.GetDatastore().Copy()
	defer sessionCopy.Close()

	coll := sessionCopy.DB(conf.SystemConfig.DBConfig.Database).C(models.COLL_NAME_SKYRING_UTILIZATION)
	if _, err := coll.Upsert(bson.M{"name": monitoring.SYSTEM}, system); err != nil {
		logger.Get().Error("%s - Error persisting the system.Error %v", ctxt, err)
	}
}
