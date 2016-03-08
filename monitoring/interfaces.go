package monitoring

type MonitoringManagerInterface interface {
	QueryDB(params map[string]interface{}) (interface{}, error)
	PushToDb(metrics map[string]map[string]string, hostName string, port int) error
	GetInstantValue(node string, resource_name string) (float64, error)
	GetResourceName(params map[string]interface{}) (string, error)
	GetInstantValuesAggregation(node string, resource_name string, exceptionResources []string) (aggregatedValue float64, err error, isCompleteFailure bool)
}
