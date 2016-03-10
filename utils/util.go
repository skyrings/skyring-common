package util

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/skyrings/skyring-common/conf"
	"github.com/skyrings/skyring-common/db"
	"github.com/skyrings/skyring-common/models"
	"github.com/skyrings/skyring-common/tools/logger"
	"github.com/skyrings/skyring-common/tools/task"
	"github.com/skyrings/skyring-common/tools/uuid"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"time"
)

// For testing, bypass HandleCrash.
var ReallyCrash bool

// PanicHandlers is a list of functions which will be invoked when a panic happens.
var PanicHandlers = []func(interface{}){logPanic}

// HandleCrash simply catches a crash and logs an error. Meant to be called via defer.
func HandleCrash() {
	if ReallyCrash {
		return
	}
	if r := recover(); r != nil {
		for _, fn := range PanicHandlers {
			fn(r)
		}
	}
}

// Deprecated. Please use Until and pass NeverStop as the stopCh.
func Forever(f func(), period time.Duration) {
	Until(f, period, nil)
}

// Until loops until stop channel is closed, running f every period.
// Catches any panics, and keeps going. f may not be invoked if
// stop channel is already closed.
func Until(f func(), period time.Duration, stopCh <-chan struct{}) {
	for {
		select {
		case <-stopCh:
			return
		default:
		}
		func() {
			defer HandleCrash()
			f()
		}()
		time.Sleep(period)
	}
}

// logPanic logs the caller tree when a panic occurs.
func logPanic(r interface{}) {
	callers := ""
	for i := 0; true; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		callers = callers + fmt.Sprintf("%v:%v\n", file, line)
	}
	logger.Get().Error("Recovered from panic: %#v (%v)\n%v", r, r, callers)
}

func FailTask(msg string, err error, t *task.Task) {
	logger.Get().Error("%s: %v", msg, err)
	t.UpdateStatus("Failed. error: %v", err)
	t.Done(models.TASK_STATUS_FAILURE)
}

func GetString(param interface{}) (string, error) {
	if str, ok := param.(string); ok {
		return str, nil
	}
	return "", fmt.Errorf("Not a string")
}

func HTTPGet(url string) ([]byte, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		return contents, nil
	}
}

func Md5FromString(name string) string {
	hash := md5.Sum([]byte(name))
	return hex.EncodeToString(hash[:])
}

func GetMapKeys(inMap interface{}) ([]reflect.Value, error) {
	inMapValue := reflect.ValueOf(inMap)
	if inMapValue.Kind() != reflect.Map {
		return nil, fmt.Errorf("Not a map!!")
	}
	keys := inMapValue.MapKeys()
	return keys, nil
}

func Stringify(keys []reflect.Value) (strings []string) {
	strings = make([]string, len(keys))
	for index, key := range keys {
		strings[index] = key.String()
	}
	return strings
}

func StringifyInterface(keys []interface{}) ([]string, error) {
	var stringKeys []string
	stringKeys = make([]string, len(keys))
	var err error
	var errString string
	for index, key := range keys {
		stringKeys[index], err = GetString(key)
		if err != nil {
			errString = errString + err.Error() + "\n"
		}
	}
	if errString != "" {
		errString = strings.TrimSpace(errString)
		return stringKeys, fmt.Errorf(errString)
	}
	return stringKeys, nil
}

func GenerifyStringArr(keys []string) []interface{} {
	iKeys := make([]interface{}, len(keys))
	for index, key := range keys {
		iKeys[index] = key
	}
	return iKeys
}

func StringSetDiff(keys1 []string, keys2 []string) (diff []string) {
	var unique bool
	for _, key1 := range keys1 {
		unique = true
		for _, key2 := range keys2 {
			if key1 == key2 {
				unique = false
				continue
			}
		}
		if unique {
			diff = append(diff, key1)
		}
	}
	return diff
}

func StringInSlice(value string, slice []string) bool {
	for _, item := range slice {
		if value == item {
			return true
		}
	}
	return false
}

func GetNodes(clusterNodes []models.ClusterNode) (map[uuid.UUID]models.Node, error) {
	sessionCopy := db.GetDatastore().Copy()
	defer sessionCopy.Close()
	var nodes = make(map[uuid.UUID]models.Node)
	coll := sessionCopy.DB(conf.SystemConfig.DBConfig.Database).C(models.COLL_NAME_STORAGE_NODES)
	for _, clusterNode := range clusterNodes {
		uuid, err := uuid.Parse(clusterNode.NodeId)
		if err != nil {
			return nodes, errors.New(fmt.Sprintf("Error parsing node id: %v", clusterNode.NodeId))
		}
		var node models.Node
		if err := coll.Find(bson.M{"nodeid": *uuid}).One(&node); err != nil {
			return nodes, err
		}
		nodes[node.NodeId] = node
	}
	return nodes, nil
}

func GetNodesByIdStr(clusterNodes []models.ClusterNode) (map[string]models.Node, error) {
	sessionCopy := db.GetDatastore().Copy()
	defer sessionCopy.Close()
	var nodes = make(map[string]models.Node)
	coll := sessionCopy.DB(conf.SystemConfig.DBConfig.Database).C(models.COLL_NAME_STORAGE_NODES)
	for _, clusterNode := range clusterNodes {
		uuid, err := uuid.Parse(clusterNode.NodeId)
		if err != nil {
			return nodes, errors.New(fmt.Sprintf("Error parsing node id: %v", clusterNode.NodeId))
		}
		var node models.Node
		if err := coll.Find(bson.M{"nodeid": *uuid}).One(&node); err != nil {
			return nodes, err
		}
		nodes[clusterNode.NodeId] = node
	}
	return nodes, nil
}
