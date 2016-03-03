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

package lock

import (
	"errors"
	"fmt"
	"github.com/skyrings/skyring-common/tools/logger"
	"github.com/skyrings/skyring-common/tools/uuid"
	"sync"
)

type Manager struct {
	/*
	   TODO: This part will move into the DB to operate between core
	   and multiple providers
	*/
	locks map[uuid.UUID]*LockInternal
}

type LockManager interface {
	//The following method will try to acquire provided lock
	AcquireLock(ctxt string, appLock AppLock) error
	// The following method will release a lock
	ReleaseLock(ctxt string, appLock AppLock)
	//The following method will clear all inserted locks
	Clear()
}

var lockMutex sync.Mutex

func NewLockManager() *Manager {
	return &Manager{make(map[uuid.UUID]*LockInternal)}
}

func (manager *Manager) AcquireLock(ctxt string, appLock AppLock) error {
	lockMutex.Lock()
	defer lockMutex.Unlock()
	//Check lock can be acquired for all the nodes. if not
	//return error
	for k := range appLock.GetAppLocks() {
		//check if the lock exists
		if val, ok := manager.locks[k]; ok {
			//Lock already aquired return from here
			err := fmt.Sprintf("Unable to Acquire the lock for %v Message %s ", k, val.GetMessages())
			logger.Get().Error("%s-Unable to Acquire the lock for: %v", ctxt, k)
			return errors.New(err)
		}
	}
	//lock can be acquired, so lock the nodes
	logger.Get().Debug("%s-Currently Locked: %v", ctxt, manager.locks)
	logger.Get().Debug("%s-Acquiring the locks for: %v", ctxt, appLock.GetAppLocks())
	for k, v := range appLock.GetAppLocks() {
		logger.Get().Debug("%s-Lock Acquired for: %v", ctxt, k)
		manager.locks[k] = NewLockInternal(v)
	}
	logger.Get().Debug("%s-Currently Locked: %v", ctxt, manager.locks)
	return nil
}

func (manager *Manager) ReleaseLock(ctxt string, appLock AppLock) {
	lockMutex.Lock()
	defer lockMutex.Unlock()
	logger.Get().Debug("%s-Currently Locked: %v", ctxt, manager.locks)
	logger.Get().Debug("%s-Releasing the locks for: %v", ctxt, appLock.GetAppLocks())
	for k := range appLock.GetAppLocks() {
		//check if the lock exists
		if _, ok := manager.locks[k]; !ok {
			//No lock exists log and do nothing
			logger.Get().Error("%s-No Lock found for unlocking: %v", ctxt, k)
		}
		logger.Get().Debug("%s-Lock Released: %v", ctxt, k)
		delete(manager.locks, k)
	}
	logger.Get().Debug("%s-Currently Locked: %v", ctxt, manager.locks)
}

func (manager *Manager) Clear() {
	lockMutex.Lock()
	defer lockMutex.Unlock()
	for k := range manager.locks {
		delete(manager.locks, k)
	}
}
