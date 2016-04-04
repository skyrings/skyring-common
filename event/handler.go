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

func ClearCorrespondingAlert(event models.AppEvent, ctxt string) (models.AlarmStatus, error) {
	var events []models.AppEvent
	sessionCopy := db.GetDatastore().Copy()
	defer sessionCopy.Close()
	coll := sessionCopy.DB(conf.SystemConfig.DBConfig.Database).C(models.COLL_NAME_APP_EVENTS)
	count := 0
	for count < 3 {
		if err := coll.Find(bson.M{
			"name":      event.Name,
			"entityid":  event.EntityId,
			"clusterid": event.ClusterId}).Sort("-timestamp").All(&events); err != nil {
			logger.Get().Error("%s-Error getting record from DB: %v", ctxt, err)
			return models.ALARM_STATUS_INDETERMINATE, err
		}
		if len(events) == 0 || events[0].Severity == models.ALARM_STATUS_CLEARED {
			if event.Severity != models.ALARM_STATUS_CLEARED {
				break
			}
			time.Sleep(180 * time.Second)
			count = count + 1
		} else {
			break
		}
	}
	if len(events) == 0 || events[0].Severity == models.ALARM_STATUS_CLEARED {
		return models.ALARM_STATUS_INDETERMINATE, errors.New(fmt.Sprintf("Corresponding alert not available for event: %s",
			event.EventId))
	} else {
		events[0].Acked = true
		events[0].SystemAckedTime = time.Now()
		events[0].AckedByEvent = event.EventId.String()
		events[0].SystemAckComment = fmt.Sprintf("This event is dismissed automatically,"+
			" as we have recieved a corresponding event: %s", event.EventId.String())

		if err := coll.Update(bson.M{"eventid": events[0].EventId},
			bson.M{"$set": events[0]}); err != nil {
			logger.Get().Warning(fmt.Sprintf("%s-Error updating record in DB for event:"+
				" %v. error: %v", ctxt, events[0].EventId.String(), err))
			return models.ALARM_STATUS_INDETERMINATE, errors.New(fmt.Sprintf("%s-Error updating"+
				" record in DB for event: %v. error: %v",
				ctxt, events[0].EventId.String(), err))
		}

	}
	return events[0].Severity, nil
}

func UpdateNodeAlarmCount(event models.AppEvent, clearedSeverity models.AlarmStatus, ctxt string) error {
	var node models.Node
	sessionCopy := db.GetDatastore().Copy()
	defer sessionCopy.Close()
	coll := sessionCopy.DB(conf.SystemConfig.DBConfig.Database).C(models.COLL_NAME_STORAGE_NODES)
	if err := coll.Find(bson.M{
		"nodeid": event.NodeId, "clusterid": event.ClusterId}).One(&node); err != nil {
		logger.Get().Error("%s-Error getting record from DB: %v", ctxt, err)
		return err
	}

	if event.Severity == models.ALARM_STATUS_MAJOR ||
		event.Severity == models.ALARM_STATUS_CRITICAL {
		node.AlmCritCount += 1
	} else if event.Severity == models.ALARM_STATUS_MINOR ||
		event.Severity == models.ALARM_STATUS_WARNING {
		node.AlmWarnCount += 1
	}

	if clearedSeverity == models.ALARM_STATUS_MAJOR ||
		clearedSeverity == models.ALARM_STATUS_CRITICAL {
		if node.AlmCritCount > 0 {
			node.AlmCritCount -= 1
		}
	} else if clearedSeverity == models.ALARM_STATUS_MINOR ||
		clearedSeverity == models.ALARM_STATUS_WARNING {
		if node.AlmWarnCount > 0 {
			node.AlmWarnCount -= 1
		}
	}

	if node.AlmCritCount > 0 {
		node.AlmStatus = models.ALARM_STATUS_CRITICAL
	} else if node.AlmWarnCount > 0 {
		node.AlmStatus = models.ALARM_STATUS_WARNING
	} else {
		node.AlmStatus = models.ALARM_STATUS_CLEARED
	}

	if err := coll.Update(bson.M{"nodeid": event.NodeId},
		bson.M{"$set": node}); err != nil {
		logger.Get().Error("%s-Error Updating the Alarm state/count: %v", ctxt, err)
		return err
	}
	return nil
}

func UpdateClusterAlarmCount(event models.AppEvent, clearedSeverity models.AlarmStatus, ctxt string) error {
	var cluster models.Cluster
	sessionCopy := db.GetDatastore().Copy()
	defer sessionCopy.Close()
	coll := sessionCopy.DB(conf.SystemConfig.DBConfig.Database).C(models.COLL_NAME_STORAGE_CLUSTERS)
	if err := coll.Find(bson.M{
		"clusterid": event.ClusterId}).One(&cluster); err != nil {
		logger.Get().Error("%s-Error getting record from DB: %v", ctxt, err)
		return err
	}

	if event.Severity == models.ALARM_STATUS_MAJOR ||
		event.Severity == models.ALARM_STATUS_CRITICAL {
		cluster.AlmCritCount += 1
	} else if event.Severity == models.ALARM_STATUS_MINOR ||
		event.Severity == models.ALARM_STATUS_WARNING {
		cluster.AlmWarnCount += 1
	}

	if clearedSeverity == models.ALARM_STATUS_MAJOR ||
		clearedSeverity == models.ALARM_STATUS_CRITICAL {
		if cluster.AlmCritCount > 0 {
			cluster.AlmCritCount -= 1
		}
	} else if clearedSeverity == models.ALARM_STATUS_MINOR ||
		clearedSeverity == models.ALARM_STATUS_WARNING {
		if cluster.AlmWarnCount > 0 {
			cluster.AlmWarnCount -= 1
		}
	}

	if cluster.AlmCritCount > 0 {
		cluster.AlmStatus = models.ALARM_STATUS_CRITICAL
	} else if cluster.AlmWarnCount > 0 {
		cluster.AlmStatus = models.ALARM_STATUS_WARNING
	} else {
		cluster.AlmStatus = models.ALARM_STATUS_CLEARED
	}

	if err := coll.Update(bson.M{"clusterid": event.ClusterId},
		bson.M{"$set": cluster}); err != nil {
		logger.Get().Error("%s-Error Updating the Alarm state/count: %v", ctxt, err)
		return err
	}
	return nil
}
