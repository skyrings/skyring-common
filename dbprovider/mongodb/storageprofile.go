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
	"github.com/skyrings/skyring-common/models"
	"github.com/skyrings/skyring-common/tools/logger"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func (m MongoDb) StorageProfile(ctxt string, name string) (sProfile models.StorageProfile, e error) {

	c := m.Connect(models.COLL_NAME_STORAGE_PROFILE)
	defer m.Close(c)
	err := c.Find(bson.M{"name": name}).One(&sProfile)
	if err != nil {
		logger.Get().Error("%s-Error getting record from DB:%s", ctxt, err)
		return sProfile, mkmgoerror(err.Error())
	}
	return sProfile, nil
}

func (m MongoDb) StorageProfiles(ctxt string, filter interface{}, ops models.QueryOps) (sProfiles []models.StorageProfile, e error) {

	c := m.Connect(models.COLL_NAME_STORAGE_PROFILE)
	defer m.Close(c)

	err := c.Find(filter).Sort("priority").Select(ops.Select).All(&sProfiles)
	if err != nil {
		logger.Get().Error("%s-Error getting record from DB:%s", ctxt, err)
		return sProfiles, mkmgoerror(err.Error())
	}
	return sProfiles, nil

}

func (m MongoDb) SaveStorageProfile(ctxt string, s models.StorageProfile) error {
	c := m.Connect(models.COLL_NAME_STORAGE_PROFILE)
	defer m.Close(c)

	_, err := c.Upsert(bson.M{"name": s.Name}, bson.M{"$set": s})
	if err != nil {
		logger.Get().Error("%s-Error deleting record from DB:%s", ctxt, err)
		return mkmgoerror(err.Error())
	}
	return nil

}
func (m MongoDb) DeleteStorageProfile(ctxt string, name string) error {
	c := m.Connect(models.COLL_NAME_STORAGE_PROFILE)
	defer m.Close(c)

	err := c.Remove(bson.M{"name": name})
	if err != nil {
		logger.Get().Error("%s-Error deleting record from DB:%s", ctxt, err)
		return mkmgoerror(err.Error())
	}
	return nil
}

func (m MongoDb) InitStorageProfile(ctxt string) error {
	//Set the indexes for storage profiles
	s := m.Connect(models.COLL_NAME_STORAGE_PROFILE)
	defer s.Database.Session.Close()

	profileIndex := mgo.Index{
		Key:    []string{"name"},
		Unique: true,
	}
	err := s.EnsureIndex(profileIndex)
	if err != nil {
		logger.Get().Error("%s-Error Initializing storage profile:%s", ctxt, err)
		return mkmgoerror(err.Error())
	}
	return nil
}
