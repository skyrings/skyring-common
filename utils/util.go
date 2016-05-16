package util

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/skyrings/skyring-common/conf"
	"github.com/skyrings/skyring-common/db"
	"github.com/skyrings/skyring-common/models"
	"github.com/skyrings/skyring-common/monitoring"
	"github.com/skyrings/skyring-common/tools/logger"
	"github.com/skyrings/skyring-common/tools/task"
	"github.com/skyrings/skyring-common/tools/uuid"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"net/http"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// For testing, bypass HandleCrash.
var ReallyCrash bool

// PanicHandlers is a list of functions which will be invoked when a panic happens.
var PanicHandlers = []func(interface{}){logPanic}

// HandleCrash simply catches a crash and logs an error. Meant to be called via defer.
func HandleCrash() {
	if ReallyCrash {
		return
	}
	if r := recover(); r != nil {
		for _, fn := range PanicHandlers {
			fn(r)
		}
	}
}

// Deprecated. Please use Until and pass NeverStop as the stopCh.
func Forever(f func(), period time.Duration) {
	Until(f, period, nil)
}

// Until loops until stop channel is closed, running f every period.
// Catches any panics, and keeps going. f may not be invoked if
// stop channel is already closed.
func Until(f func(), period time.Duration, stopCh <-chan struct{}) {
	for {
		select {
		case <-stopCh:
			return
		default:
		}
		func() {
			defer HandleCrash()
			f()
		}()
		time.Sleep(period)
	}
}

// logPanic logs the caller tree when a panic occurs.
func logPanic(r interface{}) {
	callers := ""
	for i := 0; true; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		callers = callers + fmt.Sprintf("%v:%v\n", file, line)
	}
	logger.Get().Error("Recovered from panic: %#v (%v)\n%v", r, r, callers)
}

func FailTask(msg string, err error, t *task.Task) {
	logger.Get().Error("%s: %v", msg, err)
	t.UpdateStatus("Failed. error: %v", err)
	t.Done(models.TASK_STATUS_FAILURE)
}

func GetString(param interface{}) (string, error) {
	if str, ok := param.(string); ok {
		return str, nil
	}
	return "", fmt.Errorf("Not a string")
}

func HTTPGet(url string) ([]byte, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		return contents, nil
	}
}

func Md5FromString(name string) string {
	hash := md5.Sum([]byte(name))
	return hex.EncodeToString(hash[:])
}

func GetMapKeys(inMap interface{}) ([]reflect.Value, error) {
	inMapValue := reflect.ValueOf(inMap)
	if inMapValue.Kind() != reflect.Map {
		return nil, fmt.Errorf("Not a map!!")
	}
	keys := inMapValue.MapKeys()
	return keys, nil
}

func Stringify(keys []reflect.Value) (strings []string) {
	strings = make([]string, len(keys))
	for index, key := range keys {
		strings[index] = key.String()
	}
	return strings
}

func StringifyInterface(keys []interface{}) ([]string, error) {
	var stringKeys []string
	stringKeys = make([]string, len(keys))
	var err error
	var errString string
	for index, key := range keys {
		stringKeys[index], err = GetString(key)
		if err != nil {
			errString = errString + err.Error() + "\n"
		}
	}
	if errString != "" {
		errString = strings.TrimSpace(errString)
		return stringKeys, fmt.Errorf(errString)
	}
	return stringKeys, nil
}

func GenerifyStringArr(keys []string) []interface{} {
	iKeys := make([]interface{}, len(keys))
	for index, key := range keys {
		iKeys[index] = key
	}
	return iKeys
}

func StringSetDiff(keys1 []string, keys2 []string) (diff []string) {
	var unique bool
	for _, key1 := range keys1 {
		unique = true
		for _, key2 := range keys2 {
			if key1 == key2 {
				unique = false
				continue
			}
		}
		if unique {
			diff = append(diff, key1)
		}
	}
	return diff
}

func StringInSlice(value string, slice []string) bool {
	for _, item := range slice {
		if value == item {
			return true
		}
	}
	return false
}

func GetNodes(clusterNodes []models.ClusterNode) (map[uuid.UUID]models.Node, error) {
	sessionCopy := db.GetDatastore().Copy()
	defer sessionCopy.Close()
	var nodes = make(map[uuid.UUID]models.Node)
	coll := sessionCopy.DB(conf.SystemConfig.DBConfig.Database).C(models.COLL_NAME_STORAGE_NODES)
	for _, clusterNode := range clusterNodes {
		uuid, err := uuid.Parse(clusterNode.NodeId)
		if err != nil {
			return nodes, errors.New(fmt.Sprintf("Error parsing node id: %v", clusterNode.NodeId))
		}
		var node models.Node
		if err := coll.Find(bson.M{"nodeid": *uuid}).One(&node); err != nil {
			return nodes, err
		}
		nodes[node.NodeId] = node
	}
	return nodes, nil
}

func UpdateThresholdInfoToTable(tEvent models.ThresholdEvent) (err error, isRaiseEvent bool) {
	sessionCopy := db.GetDatastore().Copy()
	defer sessionCopy.Close()
	collection := sessionCopy.DB(conf.SystemConfig.DBConfig.Database).C(models.COLL_NAME_THRESHOLD_BREACHES)
	var tEventInDb models.ThresholdEvent
	selectCriteria := bson.M{
		"entityid":        tEvent.EntityId,
		"clusterid":       tEvent.ClusterId,
		"utilizationtype": tEvent.UtilizationType,
		"entityname":      tEvent.EntityName,
	}
	err = collection.Find(selectCriteria).One(&tEventInDb)
	if err != nil {
		if err == mgo.ErrNotFound {
			if tEvent.ThresholdSeverity == strings.ToUpper(models.STATUS_OK) {
				return nil, false
			}
			if err := collection.Insert(tEvent); err != nil {
				return fmt.Errorf("Failed to persist event %v to db.Error %v",
					tEvent, err), true
			}
			return nil, true
		}
	} else {
		if tEvent.ThresholdSeverity == tEventInDb.ThresholdSeverity {
			return nil, false
		} else if tEvent.ThresholdSeverity == models.STATUS_OK {
			if err = collection.Remove(selectCriteria); err != nil {
				return fmt.Errorf("Failed to update the db, that %v of %v is back to normal.Err %v",
					tEvent.UtilizationType, tEvent.EntityName, err), true
			}
			return nil, true
		} else {
			if err := collection.Update(selectCriteria, tEvent); err != nil {
				return fmt.Errorf("Failed to persist event %v to db.Error %v",
					tEvent, err), true
			}
			return nil, true
		}
	}
	return nil, false
}

func AnalyseThresholdBreach(ctxt string, utilizationType string, resourceName string, resourceUtilization float64, cluster models.Cluster) (models.Event, bool, error) {
	var event models.Event
	timeStamp := time.Now()
	pluginIndex := monitoring.GetPluginIndex(utilizationType, cluster.Monitoring.Plugins)
	if pluginIndex == -1 {
		logger.Get().Error(
			"%s - The cluster %s is not configured to monitor threshold breaches for %v and hence cannot monitor utilization of %v",
			ctxt, cluster.Name, utilizationType, resourceName)
		return models.Event{}, false,
			fmt.Errorf(
				"%s - The cluster %s is not configured to monitor threshold breaches for %v and hence cannot monitor utilization of %v",
				ctxt, cluster.Name, utilizationType, resourceName)
	}
	plugin := cluster.Monitoring.Plugins[pluginIndex]
	var applicableConfig monitoring.PluginConfig
	var applicableThresholdValue float64
	var message string

	var entityIdentifier string
	entityId, entityIdFetchError := getEntityIdFromNameAndUtilizationType(utilizationType, resourceName, cluster)
	if entityIdFetchError != nil {
		logger.Get().Error("%s - Error fetching the id for %v in cluster %v",
			ctxt, resourceName, cluster.Name)
		return models.Event{}, false, fmt.Errorf("%s - Error fetching the id for %v in cluster %v",
			ctxt, resourceName, cluster.Name)
	}
	entityIdentifier = (*entityId).String()

	if len(plugin.Configs) == 0 {
		return models.Event{}, false, fmt.Errorf("%s - No threshold configurations found for %v of %v in cluster %v",
			ctxt, utilizationType, resourceName, cluster.Name)
	}

	for _, config := range plugin.Configs {
		if config.Category == monitoring.THRESHOLD {
			currentThresholdValue, thresholdValError := strconv.ParseFloat(config.Value, 64)
			if thresholdValError != nil {
				logger.Get().Error("%s - %v-%v configuration for %v in cluster %v could not be converted.Err %v\n",
					ctxt, config.Category, config.Type, plugin.Name, cluster.Name, thresholdValError)
			}

			if resourceUtilization > currentThresholdValue && applicableThresholdValue < currentThresholdValue {
				applicableConfig = config
				applicableThresholdValue = currentThresholdValue
			}
		}
	}

	if applicableConfig.Type == "" {
		applicableConfig.Type = models.STATUS_OK
		message = fmt.Sprintf("%v of cluster %v with value %v back to normal",
			plugin.Name, cluster.Name, resourceUtilization)
	}

	tEvent := models.ThresholdEvent{
		ClusterId:         cluster.ClusterId,
		UtilizationType:   plugin.Name,
		ThresholdSeverity: strings.ToUpper(applicableConfig.Type),
		TimeStamp:         timeStamp,
		EntityId:          *entityId,
		EntityName:        resourceName,
	}

	err, isRaiseEvent := UpdateThresholdInfoToTable(tEvent)
	if err != nil {
		logger.Get().Error("%s - %v", ctxt, err)
	}

	if isRaiseEvent {
		event = models.Event{
			Timestamp: timeStamp,
			ClusterId: cluster.ClusterId,
			NodeId:    cluster.ClusterId,
			Tag: fmt.Sprintf("skyring/%v/cluster/%v/threshold/%v/%v",
				cluster.Type, cluster.ClusterId, plugin.Name, strings.ToUpper(applicableConfig.Type)),
			Tags: map[string]string{
				models.CURRENT_VALUE:   strconv.FormatFloat(resourceUtilization, 'E', -1, 64),
				models.ENTITY_ID:       entityIdentifier,
				models.PLUGIN:          plugin.Name,
				models.THRESHOLD_TYPE:  strings.ToUpper(applicableConfig.Type),
				models.THRESHOLD_VALUE: strconv.FormatFloat(applicableThresholdValue, 'E', -1, 64),
				models.ENTITY_NAME:     resourceName,
			},
			Severity: strings.ToUpper(applicableConfig.Type),
		}
		if message == "" {
			message = fmt.Sprintf("%v of cluster %v with value %v breached %v threshold of value %v",
				plugin.Name, cluster.Name, resourceUtilization, strings.ToUpper(applicableConfig.Type), applicableThresholdValue)
		}
		event.Message = message
		return event, true, nil
	}
	return models.Event{}, false, nil
}

func getEntityIdFromNameAndUtilizationType(utilizationType string, resourceName string, cluster models.Cluster) (*uuid.UUID, error) {
	switch utilizationType {
	case monitoring.CLUSTER_UTILIZATION:
		return &(cluster.ClusterId), nil
	case monitoring.SLU_UTILIZATION:
		var slu models.StorageLogicalUnit
		err := getEntity(cluster.ClusterId, resourceName, models.COLL_NAME_STORAGE_LOGICAL_UNITS, &slu)
		if err != nil {
			return nil, fmt.Errorf("Could not fetch osd with name %v in cluster %v.Error %v",
				resourceName, cluster.Name, err)
		}
		return &(slu.SluId), nil
	case monitoring.STORAGE_UTILIZATION:
		var storage models.Storage
		err := getEntity(cluster.ClusterId, resourceName, models.COLL_NAME_STORAGE, &storage)
		if err != nil {
			return nil, fmt.Errorf("Could not fetch pool with name %v in cluster %v",
				storage.Name, cluster.Name)
		}
		return &(storage.StorageId), nil
	case monitoring.STORAGE_PROFILE_UTILIZATION:
		return &(cluster.ClusterId), nil
	case monitoring.BLOCK_DEVICE_UTILIZATION:
		var bDevice models.BlockDevice
		err := getEntity(cluster.ClusterId, resourceName, models.COLL_NAME_BLOCK_DEVICES, &bDevice)
		if err != nil {
			return nil, fmt.Errorf("Could not fetch block device with name %v in cluster %v.Error %v",
				resourceName, cluster.Name, err)
		}
		return &(bDevice.Id), nil
	}
	return nil, fmt.Errorf("Unsupported utilization type %v", utilizationType)
}

func getEntity(clusterId uuid.UUID, resourceName string, collectionName string, entity interface{}) error {
	sessionCopy := db.GetDatastore().Copy()
	defer sessionCopy.Close()
	collection := sessionCopy.DB(conf.SystemConfig.DBConfig.Database).C(collectionName)
	if err := collection.Find(bson.M{"clusterid": clusterId, "name": resourceName}).One(entity); err != nil {
		return err
	}
	return nil
}

func GetNodesByIdStr(clusterNodes []models.ClusterNode) (map[string]models.Node, error) {
	sessionCopy := db.GetDatastore().Copy()
	defer sessionCopy.Close()
	var nodes = make(map[string]models.Node)
	coll := sessionCopy.DB(conf.SystemConfig.DBConfig.Database).C(models.COLL_NAME_STORAGE_NODES)
	for _, clusterNode := range clusterNodes {
		uuid, err := uuid.Parse(clusterNode.NodeId)
		if err != nil {
			return nodes, errors.New(fmt.Sprintf("Error parsing node id: %v", clusterNode.NodeId))
		}
		var node models.Node
		if err := coll.Find(bson.M{"nodeid": *uuid}).One(&node); err != nil {
			return nodes, err
		}
		nodes[clusterNode.NodeId] = node
	}
	return nodes, nil
}

func GetReadableFloat(str string, ctxt string) (string, error) {
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		logger.Get().Error("%s-Could not parse the string: %s", ctxt, str)
		return str, err
	}
	return fmt.Sprintf("%.2f", f), nil
}
