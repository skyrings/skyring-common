package graphitemanager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/marpaia/graphite-golang"
	"github.com/skyrings/skyring-common/conf"
	"github.com/skyrings/skyring-common/monitoring"
	"github.com/skyrings/skyring-common/tools/logger"
	"github.com/skyrings/skyring-common/utils"
	"io"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"
)

const (
	TimeSeriesDBManagerName = "GraphiteManager"
)

type GraphiteManager struct {
}

func init() {
	monitoring.RegisterMonitoringManager(TimeSeriesDBManagerName, func(config io.Reader) (monitoring.MonitoringManagerInterface, error) {
		return NewGraphiteManager(config)
	})
}

func NewGraphiteManager(config io.Reader) (*GraphiteManager, error) {
	return &GraphiteManager{}, nil
}

var (
	templateUrl             = "http://{{.hostname}}:{{.port}}/render?target={{.target}}&format=json"
	targetLatest            = "cactiStyle({{.targetGeneral}})"
	targetGeneral           = "{{.collectionname}}.{{.nodename}}.{{.resourcename}}"
	targetWithParent        = "{{.collectionname}}.{{.parentname}}.{{.resourcename}}_{{.nodename}}"
	targetWildCard          = "*.*"
	templateFromTime        = "&from={{.time}}"
	templateUntilTime       = "&until={{.time}}"
	timeFormat              = "15:04_20060102"
	standardTimeFormat      = "2006-01-02T15:04:05.000Z"
	fullQualifiedMetricName bool
)

var SupportedInputTimeFormats = []string{
	"2006-01-02T15:04:05.000Z",
	"2006-01-02",
	"20060102",
}

type GraphiteMetric struct {
	Target string `json:"target"`
	Stats  stats  `json:"datapoints"`
}

type stat []interface{}
type stats []stat

var nodeNameExceptions = map[string]string{}

func (slice stats) Len() int {
	return len(slice)
}

func (slice stats) Less(i, j int) bool {
	s1, _ := slice[i][1].(int32)
	s2, _ := slice[j][1].(int32)
	return s1 < s2
}

func (slice stats) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

type GraphiteMetrics []GraphiteMetric

func formatDate(inDate string) (string, error) {
	for _, currentFormat := range SupportedInputTimeFormats {
		if parsedTime, err := time.Parse(currentFormat, inDate); err == nil {
			return parsedTime.Format("15:04") + "_" + parsedTime.Format("20060102"), nil
		}
	}
	if matched, _ := regexp.Match("^-([0-5]?[0-9])?s$|^-([0-5]?[0-9])?min$|^-([0-2]?[0-3])?h$|^-([0-9])*d$|^-([0-9])*w$|^-([0-9])*mon$|^-([0-9])*y$", []byte(inDate)); !matched {
		return "", fmt.Errorf("The date %v is of unsupported type", inDate)
	}
	return inDate, nil
}

func GetTemplateParsedString(urlParams map[string]interface{}, templateString string) (string, error) {
	buf := new(bytes.Buffer)
	parsedTemplate, err := template.New("graphite_url_parser").Parse(templateString)
	if err != nil {
		return "", err
	}
	parsedTemplate.Execute(buf, urlParams)
	return buf.String(), nil
}

var resourceCollectionNameMapper = map[string]string{
	monitoring.NETWORK_LATENCY: "ping.ping-{{.serverName}}",
	monitoring.CPU_USER:        "cpu.percent-user",
	monitoring.AGGREGATION + monitoring.INTERFACE + monitoring.OCTETS + monitoring.RX: "interface*.if_octets.rx",
	monitoring.AGGREGATION + monitoring.INTERFACE + monitoring.OCTETS + monitoring.TX: "interface*.if_octets.tx",
	monitoring.AGGREGATION + monitoring.DISK + monitoring.READ:                        "disk-*.disk_ops.read",
	monitoring.AGGREGATION + monitoring.DISK + monitoring.WRITE:                       "disk-*.disk_ops.write",
	monitoring.AGGREGATION + monitoring.MEMORY:                                        "aggregation-memory-sum.memory",
	monitoring.AGGREGATION + monitoring.SWAP:                                          "aggregation-swap-sum.swap",
	monitoring.AVERAGE + monitoring.INTERFACE + monitoring.USED:                       "interface-average.bytes-total_bandwidth_used",
	monitoring.AVERAGE + monitoring.INTERFACE + monitoring.TOTAL:                      "interface-average.bytes-total_bandwidth",
	monitoring.AVERAGE + monitoring.INTERFACE + monitoring.PERCENT:                    "interface-average.percent-network_utilization",
}

func (tsdbm GraphiteManager) GetResourceName(params map[string]interface{}) (string, error) {
	resource_name, ok := params["resource_name"].(string)
	if ok {
		collectionNameTemplate := resourceCollectionNameMapper[resource_name]
		return GetTemplateParsedString(params, collectionNameTemplate)
	} else {
		return "", fmt.Errorf("Resource %v not found", params["resource_name"])
	}
}

func getUrlBaseTemplateParams(params map[string]interface{}) (map[string]interface{}, error) {
	var nodename string
	var nodeNameError error
	if nodename, nodeNameError = util.GetString(params["nodename"]); nodeNameError != nil {
		return nil, nodeNameError
	}
	nodename = strings.Replace(nodename, ".", "_", -1)
	templateString := targetGeneral
	templateParams := map[string]interface{}{
		"collectionname": conf.SystemConfig.TimeSeriesDBConfig.CollectionName,
		"nodename":       nodename,
		"resourcename":   params["resource"],
	}
	if parentName, ok := params["parentName"]; ok {
		templateParams["parentname"] = strings.Replace(parentName.(string), ".", "_", -1)
		templateString = targetWithParent
	}
	target, targetErr := GetTemplateParsedString(templateParams, templateString)
	if targetErr != nil {
		return nil, targetErr
	}

	if !fullQualifiedMetricName {
		target = target + targetWildCard
	}

	if params["interval"] == monitoring.Latest {
		target, targetErr = GetTemplateParsedString(map[string]interface{}{"targetGeneral": target}, targetLatest)
		if targetErr != nil {
			return nil, targetErr
		}
	}
	return map[string]interface{}{
		"hostname": conf.SystemConfig.TimeSeriesDBConfig.Hostname,
		"port":     conf.SystemConfig.TimeSeriesDBConfig.Port,
		"target":   target,
	}, nil
}

func Matches(key string, keys []string) bool {
	for _, permittedKey := range keys {
		if strings.Index(key, permittedKey) == 0 {
			if key != permittedKey {
				fullQualifiedMetricName = true
			}
			return true
		}
	}
	return false
}

func ValidateRequest(params map[string]interface{}) error {
	var resource string
	if str, ok := params["resource"].(string); ok {
		resource = str
	}
	if !Matches(resource, monitoring.GeneralResources) {
		/*
			1. Ideally fetch clusterId from nodeId
			2. Fetch clustertype from clusterId
			3. Based on clusterType, delegate the querying task to appropriate provider -- bigfin etc..
				i.  The provider will check internally if its monitoring provider supports the requested resource
					a. If it supports, get the results from it
					b. Else receive the error and propogate it.
			4. Return results
		*/
		return fmt.Errorf("%v is an unsupported Resource", resource)
	}
	return nil
}

func GetRequestPath(params map[string]interface{}) (string, error) {
	urlParams, urlParamsError := getUrlBaseTemplateParams(params)
	if urlParamsError != nil {
		return "", urlParamsError
	}
	url, urlError := GetTemplateParsedString(urlParams, templateUrl)
	if urlError != nil {
		return "", urlError
	}
	if startTime, ok := params["start_time"]; ok && startTime != "" {
		validatedStartTime, validatedStartTimeError := GetValidatedGraphiteSupportedTime(startTime, templateFromTime)
		if validatedStartTimeError != nil {
			return "", validatedStartTimeError
		}
		url = url + validatedStartTime
	}

	if endTime, ok := params["end_time"]; ok && endTime != "" {
		validatedEndTime, validatedEndTimeError := GetValidatedGraphiteSupportedTime(endTime, templateUntilTime)
		if validatedEndTimeError != nil {
			return "", validatedEndTimeError
		}
		url = url + validatedEndTime
	}

	if params["interval"] != "" && params["interval"] != monitoring.Latest {
		timeString, timeStringError := util.GetString(params["interval"])
		if timeStringError != nil {
			return "", fmt.Errorf("Start time %v. Error: %v", params["start_time"], timeStringError)
		}
		dateFormat, dateFormatErr := formatDate(timeString)
		if dateFormatErr != nil {
			return "", dateFormatErr
		}
		params["interval"] = dateFormat
		var templ string
		if params["start_time"] == "" && params["end_time"] == "" {
			templ = templateFromTime
		} else if params["start_time"] != "" && params["end_time"] == "" {
			templ = templateUntilTime
		} else {
			return "", fmt.Errorf("Unsupported combination of start time %v and end time %v", params["start_time"], params["end_time"])
		}
		urlTime, urlTimeError := GetValidatedGraphiteSupportedTime(timeString, templ)
		if urlTimeError != nil {
			return "", urlTimeError
		}
		url = url + urlTime
	}
	return url, nil
}

func GetValidatedGraphiteSupportedTime(time interface{}, graphiteTimeTemplate string) (string, error) {
	timeString, timeStringError := util.GetString(time)
	if timeStringError != nil {
		return "", fmt.Errorf("Time %v. Error: %v", time, timeStringError)
	}

	dateFormat, dateFormatErr := formatDate(timeString)
	if dateFormatErr != nil {
		return "", dateFormatErr
	}

	params := make(map[string]interface{})
	params["time"] = dateFormat

	urlTime, urlTimeError := GetTemplateParsedString(params, graphiteTimeTemplate)
	if urlTimeError != nil {
		return "", urlTimeError
	}

	return urlTime, nil
}

func HttpResponse(w http.ResponseWriter, status_code int, msg string, args ...string) {
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.WriteHeader(status_code)
	var ctxt string
	if len(args) > 0 {
		ctxt = args[0]
	}
	if err := json.NewEncoder(w).Encode(msg); err != nil {
		logger.Get().Error("%s-Error: %v", ctxt, err)
	}
	return
}

func (tsdbm GraphiteManager) QueryMonitoringDB(urlStr string, w http.ResponseWriter, r *http.Request) error {
	url := fmt.Sprintf("http://%s:%s/render?%s", conf.SystemConfig.TimeSeriesDBConfig.Hostname, strconv.Itoa(conf.SystemConfig.TimeSeriesDBConfig.Port), urlStr)
	http.Redirect(w, r, url, http.StatusFound)
	return nil
}

func (tsdbm GraphiteManager) QueryDB(params map[string]interface{}) (interface{}, error) {
	err := ValidateRequest(params)
	if err != nil {
		return nil, err
	}
	url, err := GetRequestPath(params)
	if err != nil {
		return nil, err
	}
	results, getError := util.HTTPGet(url)
	if getError != nil {
		return nil, getError
	} else {
		var metrics []GraphiteMetric
		if mErr := json.Unmarshal(results, &metrics); mErr != nil {
			return nil, fmt.Errorf("Error unmarshalling the metrics %v.Error:%v", metrics, mErr)
		}
		if params["interval"] == monitoring.Latest {
			/*
					 The output of cactistyle graphite function(It gives summarised output including max value, min value and current non-nil value) is as under:
				       [{"target": "<metric_name> Current:<current_val> Max:<max> Min:<min> ", "datapoints": [[val1, timestamp1],...]}
					....
					]
			*/
			for metricIndex, metric := range metrics {
				target := metric.Target
				if !strings.Contains(target, "Current:") {
					return nil, fmt.Errorf("Current non-nil value is not available for %v of %v and the target available from time series db is %v", params["resource"], params["nodename"], target)
				}
				targetContents := strings.Fields(target)
				index := -1
				for contentIndex, content := range targetContents {
					if strings.Contains(content, "Current:") {
						index = contentIndex
						break
					}
				}
				if index == -1 {
					return nil, fmt.Errorf("Current non-nil value is not available for %v of %v and the target available from time series db is %v", params["resource"], params["nodename"], target)
				}

				statValues := strings.Split(targetContents[index], ":")
				if len(statValues[index]) == 0 {
					statValues = []string{"Current", targetContents[index+1]}
				}
				if len(statValues) != 2 {
					return nil, fmt.Errorf("Current non-nil value is not available for %v of %v and the target available from time series db is %v", params["resource"], params["nodename"], target)
				}
				var statVar []interface{}
				statVar = append(statVar, statValues[1])
				/*
				 Effectively the current value returned by graphite is actually the last non-nil
				 value in graphite for the metric
				 Since graphite doesn't return the time stamp of curr ent value it is felt safe to
				 insert a 0 for the field,
				 instead of meddling into looking at when actually the currentvalue appears in the
				 possibly humongously huge set of data values
				*/
				statVar = append(statVar, 0)
				metrics[metricIndex].Target = targetContents[0]
				metrics[metricIndex].Stats = stats([]stat{statVar})
			}
		} else {
			// Filter out all null values
			for metricIndex := range metrics {
				deleted := 0
				for statIndex := range metrics[metricIndex].Stats {
					j := statIndex - deleted
					if metrics[metricIndex].Stats[j][0] == nil {
						metrics[metricIndex].Stats = metrics[metricIndex].Stats[:j+copy(metrics[metricIndex].Stats[j:], metrics[metricIndex].Stats[j+1:])]
						deleted++
					}
				}
			}
		}
		return metrics, nil
	}
}

func (tsdbm GraphiteManager) GetInstantValue(node string, resource_name string) (float64, error) {
	node = strings.Replace(node, ".", "_", -1)
	paramsToQuery := map[string]interface{}{"nodename": node, "resource": resource_name, "interval": monitoring.Latest}
	mStats, memoryStatsFetchError := tsdbm.QueryDB(paramsToQuery)
	if memoryStatsFetchError != nil {
		return 0, fmt.Errorf("Failed to fetch instant %v statistics for %s.Error: %v", resource_name, node, memoryStatsFetchError)
	}
	metrics, ok := mStats.([]GraphiteMetric)
	if !ok || len(metrics) != 1 {
		return 0, fmt.Errorf("Failed to get the instant stat of %v resource of %v", resource_name, node)
	}
	statistics := []stat(metrics[0].Stats)
	timeStampValueArr := []interface{}(statistics[0])
	if !ok {
		return 0.0, fmt.Errorf("Failed to get the instant stat of %v resource of %v", resource_name, node)
	}
	value, ok := timeStampValueArr[0].(string)
	if !ok {
		return 0.0, fmt.Errorf("Failed to get the instant stat of %v resource of %v", resource_name, node)
	}
	fVal, fValErr := strconv.ParseFloat(value, 64)
	if fValErr != nil {
		return 0.0, fmt.Errorf("Failed to get instant stat of %v resource of %v. The value %v is not a number.Err %v", resource_name, node, value, fValErr)
	}
	if math.IsNaN(fVal) {
		return 0.0, fmt.Errorf("The value %v from resource %v of %v is not a number", fVal, resource_name, node)
	}
	return fVal, nil
}

func isSeriesException(seriesName string, exceptionResources []string, node string) bool {
	for _, eResource := range exceptionResources {
		if strings.HasPrefix(seriesName, conf.SystemConfig.TimeSeriesDBConfig.CollectionName+"."+node+"."+eResource+".") {
			return true
		}
	}
	return false
}

func (tsdbm GraphiteManager) GetInstantValuesAggregation(node string, resource_name string, exceptionResources []string) (aggregatedValue float64, err error, isCompleteFailure bool) {
	var err_str string
	node = strings.Replace(node, ".", "_", -1)
	paramsToQuery := map[string]interface{}{"nodename": node, "resource": resource_name, "interval": monitoring.Latest}
	mStats, mStatsFetchError := tsdbm.QueryDB(paramsToQuery)
	if mStatsFetchError != nil {
		return 0, fmt.Errorf("Failed to fetch instant %v statistics for %s.Error: %v", resource_name, node, mStatsFetchError), true
	}
	metrics, ok := mStats.([]GraphiteMetric)
	if !ok || len(metrics) == 0 {
		return 0, fmt.Errorf("Failed to get the instant stat of %v resource of %v", resource_name, node), true
	}
	metric_cnt := len(metrics)
	for _, metric := range metrics {
		if !isSeriesException(metric.Target, exceptionResources, node) {
			statistics := []stat(metric.Stats)
			timeStampValueArr := []interface{}(statistics[0])
			if !ok {
				err_str = err_str + fmt.Sprintf("Failed to get the instant stat of %v from %v of %v\n", resource_name, metric.Target, node)
				// Decrement metric_cnt for average as the current node is not contributing.
				metric_cnt = metric_cnt - 1
				continue
			}
			value, ok := timeStampValueArr[0].(string)
			if !ok {
				err_str = err_str + fmt.Sprintf("Failed to get the instant stat of %v from %v of %v\n", resource_name, metric.Target, node)
				// Decrement metric_cnt for average as the current node is not contributing.
				metric_cnt = metric_cnt - 1
				continue
			}
			val, err := strconv.ParseFloat(value, 64)
			if err != nil {
				err_str = err_str + fmt.Sprintf("Failed to get the instant stat of %v from %v of %v\n", resource_name, metric.Target, node)
				// Decrement metric_cnt for average as the current node is not contributing.
				metric_cnt = metric_cnt - 1
				continue
			}
			if !math.IsNaN(val) {
				aggregatedValue = aggregatedValue + val
			}
		} else {
			//The series is in exception list and need not be counted for averaging out.
			metric_cnt = metric_cnt - 1
		}
	}
	if err_str != "" {
		return aggregatedValue / float64(metric_cnt), fmt.Errorf("%v", strings.TrimSpace(err_str)), metric_cnt == 0
	}
	return aggregatedValue / float64(metric_cnt), nil, metric_cnt == 0
}

//This method takes map[string]map[string]string ==> map[metric/table name]map[timestamp]value
func (tsdbm GraphiteManager) PushToDb(metrics map[string]map[string]string, hostName string, port int) error {
	data := make([]graphite.Metric, 1)
	for tableName, valueMap := range metrics {
		for timestamp, value := range valueMap {
			timeInt, err := strconv.ParseInt(timestamp, 10, 64)
			if err != nil {
				return fmt.Errorf("Failed to parse timestamp %v of metric tableName %v.Error: %v", timestamp, tableName, err.Error())
			}
			data = append(data, graphite.Metric{Name: tableName, Value: value, Timestamp: timeInt})
		}
	}
	Graphite, err := graphite.NewGraphite(hostName, port)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	Graphite.SendMetrics(data)
	if err := Graphite.Disconnect(); err != nil {
		logger.Get().Warning("Failed to disconnect the graphite conn.Error %v", err)
	}
	return nil
}
