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
package mongodb

import (
	"errors"
	"fmt"
	"github.com/skyrings/skyring-common/models"
	"github.com/skyrings/skyring-common/tools/logger"
	"gopkg.in/mgo.v2/bson"
)

var (
	ErrMissingNotifier = mkmgoerror("can't find Mail Notifier")
)

// User returns the Mail notifier.
func (m MongoDb) MailNotifier(ctxt string) (models.MailNotifier, error) {
	c := m.Connect(models.COLL_NAME_MAIL_NOTIFIER)
	defer m.Close(c)
	var notifier []models.MailNotifier
	if err := c.Find(nil).All(&notifier); err != nil || len(notifier) == 0 {
		logger.Get().Error("%s-Unable to read MailNotifier from DB: %v", ctxt, err)
		return models.MailNotifier{}, ErrMissingNotifier
	} else {
		return notifier[0], nil
	}
}

// Save mail notifier adds a new mail notifier, it replaces the existing one if there
// is already a notifier available.
func (m MongoDb) SaveMailNotifier(ctxt string, notifier models.MailNotifier) error {
	c := m.Connect(models.COLL_NAME_MAIL_NOTIFIER)
	defer m.Close(c)
	_, err := c.Upsert(bson.M{}, bson.M{"$set": notifier})
	if err != nil {
		logger.Get().Error("%s-Error Updating the mail notifier info for: %s Error: %v", ctxt, notifier.MailId, err)
		return errors.New(fmt.Sprintf("Error Updating the mail notifier info for: %s Error: %v", notifier.MailId, err))
	}
	return nil
}
