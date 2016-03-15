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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/parnurzeal/gorequest"
	"github.com/skyrings/skyring-common/conf"
	"github.com/skyrings/skyring-common/models"
	"github.com/skyrings/skyring-common/provisioner"
	"github.com/skyrings/skyring-common/tools/logger"
	"github.com/skyrings/skyring-common/utils"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type CephInstaller struct {
}

const (
	ProviderName                      = "ceph-installer"
	CEPH_INSTALLER_API_DEFAULT_PREFIX = "api"
	CEPH_INSTALLER_API_PORT           = 8181
	MON                               = "MON"
	OSD                               = "OSD"
)

type InstallReq struct {
	Hosts         []string `json:"hosts"`
	RedhatStorage bool     `json:"redhat_storage"`
	RedhatUseCdn  bool     `json:"redhat_use_cdn"`
}

type Task struct {
	Endpoint   string `json:"endpoint"`
	Succeeded  bool   `json:"succeeded"`
	Stdout     string `json:"stdout"`
	Started    string `json:"started"`
	ExitCode   int    `json:"exit_code"`
	Ended      string `json:"ended"`
	Command    string `json:"command"`
	Stderr     string `json:"stderr"`
	Identifier string `json:"identifier"`
}

var httprequest = gorequest.New()

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

func (c CephInstaller) Install(ctxt string, nodes []models.ClusterNode) []models.ClusterNode {

	db_nodes, err := util.GetNodesByIdStr(nodes)
	if err != nil {
		logger.Get().Error("%s-Error getting nodes while package installation: %v", ctxt, err)
		return nodes
	}

	var (
		wg          sync.WaitGroup
		failedNodes []models.ClusterNode
		syncMutex   sync.Mutex
	)

	for _, node := range nodes {
		wg.Add(1)
		go func(node models.ClusterNode, hostname string) {
			defer wg.Done()
			var route models.ApiRoute
			hosts := []string{hostname}
			if util.StringInSlice(MON, node.NodeType) {
				route = CEPH_INSTALLER_API_ROUTES["monInstall"]
			} else if util.StringInSlice(OSD, node.NodeType) {
				route = CEPH_INSTALLER_API_ROUTES["osdInstall"]
			}
			/*
				TODO replace the map with commented code once the ceph-installer support
				available
			*/
			/*data := InstallReq{Hosts: hosts,
				RedhatStorage: conf.SystemConfig.Provisioners[bigfin_conf.ProviderName].RedhatStorage,
				RedhatUseCdn:  conf.SystemConfig.Provisioners[bigfin_conf.ProviderName].RedhatUseCdn}
			resp, body, errs := httprequest.Post(formUrl(route)).SendStruct(data).End()*/
			data := make(map[string]interface{})
			data["hosts"] = hosts
			data["redhat_storage"] = conf.SystemConfig.Provisioners["ceph"].RedhatStorage
			jsonString, err := json.Marshal(data)
			if err != nil {
				logger.Get().Error("%s-Error marshalling data: %v", ctxt, err)
				return
			}
			resp, body, errs := httprequest.Post(formUrl(route)).Send(string(jsonString)).End()
			respData, err := parseInstallResponseData(ctxt, resp, body, errs)
			if err != nil {
				syncMutex.Lock()
				defer syncMutex.Unlock()
				failedNodes = append(failedNodes, node)
				return
			}
			if ok := syncRequestStatus(ctxt, respData.Identifier); !ok {
				syncMutex.Lock()
				defer syncMutex.Unlock()
				failedNodes = append(failedNodes, node)
				return
			}

		}(node, db_nodes[node.NodeId].Hostname)
	}
	wg.Wait()

	return failedNodes
}
func (c CephInstaller) Configure(ctxt string, reqType string, data map[string]interface{}) error {
	var (
		resp  *http.Response
		body  string
		errs  []error
		route models.ApiRoute
	)
	if reqType == MON {
		route = CEPH_INSTALLER_API_ROUTES["monConfigure"]
		jsonString, err := json.Marshal(data)
		if err != nil {
			return err
		}
		resp, body, errs = httprequest.Post(formUrl(route)).Send(jsonString).End()

	} else if reqType == OSD {
		route = CEPH_INSTALLER_API_ROUTES["osdConfigure	"]
		jsonString, err := json.Marshal(data)
		if err != nil {
			return err
		}
		resp, body, errs = httprequest.Post(formUrl(route)).Send(jsonString).End()

	}
	respData, err := parseInstallResponseData(ctxt, resp, body, errs)
	if err != nil {
		return err
	}
	if ok := syncRequestStatus(ctxt, respData.Identifier); !ok {
		return err
	}
	return nil
}

func formUrl(route models.ApiRoute) string {
	url := fmt.Sprintf("http://localhost:%d/%s/%s", CEPH_INSTALLER_API_PORT, CEPH_INSTALLER_API_DEFAULT_PREFIX, route.Pattern)
	return url
}

func syncRequestStatus(ctxt string, taskId string) bool {
	var status bool

	for {
		time.Sleep(10 * time.Second)
		taskData, err := getTaskData(ctxt, taskId)
		if err != nil {
			return status
		}

		if taskData.Ended != "" {
			// If request has failed return with error
			if taskData.Succeeded {
				status = true
			} else {
				logger.Get().Error("%s-Error while installing Packages. stdout: %v. stderr:%v", ctxt, taskData.Stdout, taskData.Stderr)
			}
			break
		}
	}

	return status
}

func getTaskData(ctxt string, taskId string) (Task, error) {

	var respData Task
	route := CEPH_INSTALLER_API_ROUTES["taskStatus"]
	route.Pattern = strings.Replace(route.Pattern, "{id}", taskId, 1)
	resp, body, errs := httprequest.Get(formUrl(route)).End()
	if len(errs) > 0 {
		logger.Get().Error("%s-Error getting task status for task:%v.Err:%v", ctxt, taskId, errs)
		return respData, errors.New("Error getting task status")
	}
	if resp.StatusCode != http.StatusOK {
		logger.Get().Error("%s-Error Getting Task Status:%v", ctxt, taskId)
		return respData, errors.New("Error getting task status")
	}

	if err := json.Unmarshal([]byte(body), &respData); err != nil {
		logger.Get().Error("%s-Unable to unmarshal response. error: %v", ctxt, err)
		return respData, errors.New("Unable to unmarshal response")
	}
	return respData, nil
}

func parseInstallResponseData(ctxt string, resp *http.Response, body string, errs []error) (Task, error) {
	var respData Task
	if len(errs) > 0 {
		logger.Get().Error("%s-Error Installing Packages:%v", ctxt, errs)
		return respData, errors.New("Error Installing Packages")
	}
	if resp.StatusCode != http.StatusOK {
		logger.Get().Error("%s-Error Installing Packages", ctxt)
		return respData, errors.New("Error Installing Packages. Ceph-installer failed")
	}
	if err := json.Unmarshal([]byte(body), &respData); err != nil {
		logger.Get().Error("%s-Unable to unmarshal response. error: %v", ctxt, err)
		return respData, errors.New("Unable to unmarshal response")
	}
	return respData, nil
}
