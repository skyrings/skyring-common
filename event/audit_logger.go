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
	"fmt"
	"github.com/skyrings/skyring-common/conf"
	"github.com/skyrings/skyring-common/db"
	"github.com/skyrings/skyring-common/dbprovider"
	"github.com/skyrings/skyring-common/models"
	"github.com/skyrings/skyring-common/notifier"
	"github.com/skyrings/skyring-common/tools/logger"
	"gopkg.in/mgo.v2/bson"
)

func AuditLog(ctxt string, event models.AppEvent, dbprovider dbprovider.DbInterface) error {
	sessionCopy := db.GetDatastore().Copy()
	defer sessionCopy.Close()
	coll := sessionCopy.DB(conf.SystemConfig.DBConfig.Database).C(models.COLL_NAME_STORAGE_CLUSTERS)
	var cluster models.Cluster
	if err := coll.Find(bson.M{"clusterid": event.ClusterId}).One(&cluster); err == nil {
		event.ClusterName = cluster.Name
	}

	if event.Notify {
		subject, body, err := getMailDetails(event)
		if err != nil {
			logger.Get().Error("%s-Could not get mail details for event: %s", ctxt, event.Name)
		} else {
			if err = notifier.MailNotify(subject, body, dbprovider, ctxt); err == nil {
				event.Notified = true
			} else {
				logger.Get().Error("%s-Could not send mail for event: %s", ctxt, event.Name)
			}
		}
	}

	coll = sessionCopy.DB(conf.SystemConfig.DBConfig.Database).C(models.COLL_NAME_APP_EVENTS)
	if err := coll.Insert(event); err != nil {
		logger.Get().Error("%s-Error adding the app event: %v", ctxt, err)
		return err
	}
	return nil
}

func getMailDetails(event models.AppEvent) (string, string, error) {
	var subject string
	var body string
	subject = fmt.Sprintf("%s: %s", event.Severity.String(), event.Message)
	body = fmt.Sprintf("%s:\n\nTimestamp: %s\nAffected entity: %s\nOrigin node:%s\n", event.Name, event.Timestamp, event.NotificationEntity.String(), event.NodeName)
	if event.Description != "" {
		body += fmt.Sprintf("Description: %s\n", event.Description)
	}
	if event.ClusterName != "" {
		body += fmt.Sprintf("Cluster: %s\n", event.ClusterName)
	}
	for k := range event.Tags {
		body += fmt.Sprintf("%s: %s\n", k, event.Tags[k])
	}
	return subject, body, nil
}
