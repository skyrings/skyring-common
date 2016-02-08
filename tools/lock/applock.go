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
	"github.com/skyrings/skyring-common/tools/uuid"
)

type AppLock struct {
	locks map[uuid.UUID]string
}

func (a *AppLock) GetAppLocks() map[uuid.UUID]string {
	return a.locks
}

func NewAppLock(locks map[uuid.UUID]string) *AppLock {
	return &AppLock{locks}
}
