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
package event

import (
	"errors"
	"fmt"
	"github.com/skyrings/skyring-common/conf"
	"github.com/skyrings/skyring-common/db"
	"github.com/skyrings/skyring-common/models"
	"github.com/skyrings/skyring-common/tools/logger"
	"github.com/skyrings/skyring-common/tools/uuid"
	"gopkg.in/mgo.v2/bson"
	"time"
)

func Persist_event(event models.Event, ctxt string) error {
	sessionCopy := db.GetDatastore().Copy()
	defer sessionCopy.Close()
	coll := sessionCopy.DB(conf.SystemConfig.DBConfig.Database).C(models.COLL_NAME_NODE_EVENTS)
	if err := coll.Insert(event); err != nil {
		logger.Get().Error("%s-Error adding the node event: %v", ctxt, err)
		return err
	}
	return nil
}

func ClearCorrespondingAlert(event models.AppEvent, ctxt string) error {
	var events []models.AppEvent
	sessionCopy := db.GetDatastore().Copy()
	defer sessionCopy.Close()
	coll := sessionCopy.DB(conf.SystemConfig.DBConfig.Database).C(models.COLL_NAME_APP_EVENTS)
	count := 0
	for count < 5 {
		if err := coll.Find(bson.M{
			"name":      event.Name,
			"entityid":  event.EntityId,
			"clusterid": event.ClusterId}).Sort("-timestamp").All(&events); err != nil {
			logger.Get().Error("%s-Error getting record from DB: %v", ctxt, err)
			return err
		}
		if len(events) == 0 || events[0].Severity == models.ALARM_STATUS_CLEARED {
			time.Sleep(180 * time.Second)
			count = count + 1
		} else {
			break
		}
	}
	if len(events) == 0 || events[0].Severity == models.ALARM_STATUS_CLEARED {
		logger.Get().Warning("%s-Corresponding alert not available for event: %s",
			ctxt, event.EventId)
		return errors.New(fmt.Sprintf("Corresponding alert not available for event: %s",
			event.EventId))
	} else {
		events[0].Acked = true
		events[0].SystemAckedTime = time.Now()
		events[0].AckedByEvent = event.EventId.String()
		events[0].SystemAckComment = fmt.Sprintf("This event is dismissed automatically,"+
			" as we have recieved a corresponding recovery event: %s", event.EventId.String())

		if err := coll.Update(bson.M{"eventid": events[0].EventId},
			bson.M{"$set": events[0]}); err != nil {
			logger.Get().Warning(fmt.Sprintf("%s-Error updating record in DB for event:"+
				" %v. error: %v", ctxt, events[0].EventId.String(), err))
			return errors.New(fmt.Sprintf("%s-Error updating record in DB for event:"+
				" %v. error: %v", ctxt, events[0].EventId.String(), err))
		}

	}
	return nil
}

func UpdateNodeAlarmCount(entityId uuid.UUID, clusterId uuid.UUID, alarmState models.AlarmStatus, ctxt string) error {
	var node models.Node
	sessionCopy := db.GetDatastore().Copy()
	defer sessionCopy.Close()
	coll := sessionCopy.DB(conf.SystemConfig.DBConfig.Database).C(models.COLL_NAME_STORAGE_NODES)
	if err := coll.Find(bson.M{
		"nodeid": entityId, "clusterid": clusterId}).One(&node); err != nil {
		logger.Get().Error("%s-Error getting record from DB: %v", ctxt, err)
		return err
	}
	if alarmState == models.ALARM_STATUS_CLEARED && node.AlmCount > 0 {
		node.AlmCount = node.AlmCount - 1
		if node.AlmCount == 0 {
			node.AlmStatus = alarmState
		}
	} else {
		node.AlmCount = node.AlmCount + 1
		if node.AlmStatus > alarmState {
			node.AlmStatus = alarmState
		}
	}
	if err := coll.Update(bson.M{"nodeid": entityId},
		bson.M{"$set": node}); err != nil {
		logger.Get().Error("%s-Error Updating the Alarm state/count: %v", ctxt, err)
		return err
	}
	return nil
}

func UpdateClusterAlarmCount(entityId uuid.UUID, alarmState models.AlarmStatus, ctxt string) error {
	var cluster models.Cluster
	sessionCopy := db.GetDatastore().Copy()
	defer sessionCopy.Close()
	coll := sessionCopy.DB(conf.SystemConfig.DBConfig.Database).C(models.COLL_NAME_STORAGE_CLUSTERS)
	if err := coll.Find(bson.M{
		"clusterid": entityId}).One(&cluster); err != nil {
		logger.Get().Error("%s-Error getting record from DB: %v", ctxt, err)
		return err
	}
	if alarmState == models.ALARM_STATUS_CLEARED && cluster.AlmCount > 0 {
		cluster.AlmCount = cluster.AlmCount - 1
		if cluster.AlmCount == 0 {
			cluster.AlmStatus = alarmState
		}
	} else {
		cluster.AlmCount = cluster.AlmCount + 1
		if cluster.AlmStatus > alarmState {
			cluster.AlmStatus = alarmState
		}
	}
	if err := coll.Update(bson.M{"clusterid": entityId},
		bson.M{"$set": cluster}); err != nil {
		logger.Get().Error("%s-Error Updating the Alarm state/count: %v", ctxt, err)
		return err
	}
	return nil
}
