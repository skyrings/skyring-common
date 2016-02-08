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
package util

import (
	"fmt"
	"reflect"
	"strings"
)

/*
Golang doesn't provide Set type.
The easiest O(1) way of implementing one is using maps.
*/
type Set map[interface{}]bool

var elementType reflect.Type

//Use this for a type independent Set
func NewSet() Set {
	return make(map[interface{}]bool)
}

// Use this for conventional Set of a required type
func NewSetWithType(t reflect.Type) Set {
	elementType = t
	return make(map[interface{}]bool)
}

func (s *Set) AddAll(elements []interface{}) error {
	var errStr string
	for _, element := range elements {
		if err := s.Add(element); err != nil {
			errStr = errStr + err.Error() + "\n"
		}
	}
	errStr = strings.TrimSpace(errStr)
	return fmt.Errorf(errStr)
}

func (s *Set) Add(element interface{}) error {
	if elementType != nil && reflect.TypeOf(element) != elementType {
		return fmt.Errorf("Element not of desired type")
	}
	if (*s)[element] == false {
		(*s)[element] = true
		return nil
	}
	return fmt.Errorf("Element %v is duplicate", element)
}

func (s *Set) Remove(element interface{}) error {
	if (*s)[element] == false {
		return fmt.Errorf("Element %v not in set", element)
	}
	delete((*s), element)
	return nil
}

func (s Set) GetElements() (values []interface{}) {
	for key := range s {
		values = append(values, key)
	}
	return values
}
