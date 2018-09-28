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
	"encoding/json"
	"testing"
)

func TestParseJSONData(t *testing.T) {
	var msg = "[{\"target\": \"test\", \"datapoints\": [[3.5805225, 1526986610], [3.5805225, 1526986611]]}]"
	var jsonData interface{}
	unmarshalErr := json.Unmarshal([]byte(msg), &jsonData)
	if unmarshalErr != nil {
		t.Error(
			"Unmarshalling failed for", msg,
		)
	}

	var targetsData = ParseJSONData(jsonData)
	if len(targetsData) == 0 {
		t.Error(
			"Should not empty", targetsData,
		)
	}

	if len(targetsData["test"].datapoints) == 0 {
		t.Error(
			"Should not empty", targetsData["test"].datapoints,
		)
	}
}

func TestAverage(t *testing.T) {
	var msg = "[{\"target\": \"test\", \"datapoints\": [[3.5805225, 1526986610], [3.5805225, 1526986611]]}, {\"target\": \"test2\", \"datapoints\": [[3, 1526986610], [6, 1526986611], [3, 1526986612]]}]"
	var jsonData interface{}
	json.Unmarshal([]byte(msg), &jsonData)
	var targetsData = ParseJSONData(jsonData)

	if targetsData["test"].Average() != 3.5805225 {
		t.Error(
			"Should be average", targetsData["test"].Average(),
		)
	}

	if targetsData["test2"].Average() != 4 {
		t.Error(
			"Should be average", targetsData["test2"].Average(),
		)
	}
}
