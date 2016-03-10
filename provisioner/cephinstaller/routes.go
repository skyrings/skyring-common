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

package cephinstaller

import "github.com/skyrings/skyring-common/models"

var (
	CEPH_INSTALLER_API_ROUTES = make(map[string]models.ApiRoute)
)

func (c *CephInstaller) LoadRoutes() {

	var routes = []models.ApiRoute{
		{
			Name:    "monInstall",
			Pattern: "mon/install",
			Method:  "POST",
		},
		{
			Name:    "monConfigure",
			Pattern: "mon/configure",
			Method:  "POST",
		},
		{
			Name:    "osdInstall",
			Pattern: "osd/install",
			Method:  "POST",
		},
		{
			Name:    "osdConfigure",
			Pattern: "osd/configure",
			Method:  "POST",
		},
		{
			Name:    "taskStatus",
			Pattern: "tasks/{id}",
			Method:  "GET",
		},
	}

	for _, route := range routes {
		CEPH_INSTALLER_API_ROUTES[route.Name] = route
	}
}
