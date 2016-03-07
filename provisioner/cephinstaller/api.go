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

import (
	"github.com/skyrings/skyring-common/provisioner"
	"io"
)

type CephInstaller struct {
}

const (
	ProviderName = "ceph-installer"
)

func init() {
	provisioner.RegisterProvider(ProviderName, func(config io.Reader) (provisioner.Provisioner, error) {
		return New(config)
	})
}

func New(config io.Reader) (*CephInstaller, error) {
	installer := new(CephInstaller)
	installer.LoadRoutes()
	return installer, nil
}

func (c CephInstaller) Install(role string, hosts []string) error {
	return nil
}
func (c CephInstaller) Configure(role string, data map[string]interface{}) error {
	return nil
}
