/*
Copyright 2018 Lars Eric Scheidler

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package graphite provides struct and functions to get datapoints from graphite web
package graphite

import (
  "io/ioutil"
  "fmt"
  "encoding/json"
  "errors"
  "net/http"
  "path/filepath"
  "strings"
  "time"
)

// Datapoint holds a graphite datapoint returned from graphite web
type Datapoint struct {
  value float32
  timestamp int
}

// Target holds a target and its datapoints returned from graphite web
type Target struct {
  target string
  datapoints []Datapoint
}

// Average returns the average calculated over all datapoints
func (target Target) Average() float32 {
  var sum float32 = 0
  for _, datapoint := range target.datapoints {
    sum += datapoint.value
  }
  return sum / float32(len(target.datapoints))
}

// Target returns target name
func (target Target) Target() string {
  return target.target
}

// Find finds a target in data map
func Find(data map[string]Target, targetName string) (Target, bool) {
  target := Target{}
  targetOk := false

  // if targetName contains a glob, we need to match the glob with all targets available
  if strings.Contains(targetName, "*") {
    for _, targetElem := range data {
      if matched, _ := filepath.Match(targetName, targetElem.Target()); matched {
        return targetElem, true
      }
    }
  } else {
    target, targetOk = data[targetName]
  }

  return target, targetOk
}

// GetTargets gets data from graphite web and returns a slice of targets
func GetTargets(graphiteUrl string, targets []string, from string, until string, timeout time.Duration) (map[string]Target, error) {
  jsonData, err := Get(graphiteUrl, targets, from, until, timeout)
  if err != nil {
    return nil, err
  }

  return ParseJsonData(jsonData), nil
}

// ParseJsonData parses json data
func ParseJsonData(jsonData interface{}) map[string]Target {
  targetsData := map[string]Target{}
  for _, elem := range jsonData.([]interface{}) {
    elem := elem.(map[string]interface{})
    target := Target{target: elem["target"].(string), datapoints: []Datapoint{}}

    for _, datapoint_elem := range elem["datapoints"].([]interface{}) {
      datapoint_elem := datapoint_elem.([]interface{})

      // ignore datapoints, where value is nil
      if datapoint_elem[0] != nil {
        datapoint := Datapoint{timestamp: int(datapoint_elem[1].(float64)), value: float32(datapoint_elem[0].(float64))}

        target.datapoints = append(target.datapoints, datapoint)
      }
    }

    targetsData[target.target] = target
  }
  return targetsData
}

// Get gets datapoints from graphite web
func Get(graphiteUrl string, targets []string, from string, until string, timeout time.Duration) (interface{}, error) {
  client := &http.Client{
    Timeout: time.Duration(timeout * time.Second),
  }

  resp, getErr := client.Get(graphiteUrl + "/render?format=json&target=" + strings.Join(targets, "&target=") + "&from=" + from + "&until=" + until)
  if getErr != nil {
    return nil, getErr
  }
  defer resp.Body.Close()

  if resp.StatusCode != 200 {
    return nil, errors.New(fmt.Sprintf("Wrong status code: %d", resp.StatusCode))
  }

  body, bodyReadErr := ioutil.ReadAll(resp.Body)
  if bodyReadErr != nil {
    return nil, bodyReadErr
  }

  var jsonData interface{}
  unmarshalErr := json.Unmarshal(body, &jsonData)
  if unmarshalErr != nil {
    return nil, unmarshalErr
  }
  return jsonData, nil
}
