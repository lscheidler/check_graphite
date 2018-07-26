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

// Package check provides functions regarding graphite check
package check

import(
  "fmt"
  "math"
  "regexp"
  "strings"

  "github.com/lscheidler/go-nagios"
  "github.com/lscheidler/check_graphite/graphite"
)

type Check struct {
  Critical float64
  EmptyOk bool
  Name string
  nagios nagios.Nagios
  Max float64
  Percentage bool
  Perfdata bool
  SumFlag bool
  TargetFlag *Target
  TargetNameRegexp string
  Warning float64
}

type Target []string

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (i *Target) String() string {
  return fmt.Sprint(*i)
}

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
// It's a comma-separated list, so we split it.
func (i *Target) Set(value string) error {
  for _, t := range strings.Split(value, ",") {
    *i = append(*i, t)
  }
  return nil
}

func (check *Check) Exit() {
  check.nagios.Exit()
}

func (check *Check) Error(err error) {
  check.nagios.Unknown(err.Error())
}

// check checks graphite targets against thresholds
func (check *Check) runCheck(data map[string]graphite.Target) {
  check.nagios.ShowPerfdata = check.Perfdata

  if check.Percentage && math.IsNaN(check.Max) {
    check.checkPercentage(data)
  } else {
    sum := check.checkTargets(data)

    if check.SumFlag {
      if check.Percentage && ! math.IsNaN(check.Max) {
        check.nagios.CheckPercentageThreshold(check.Name, sum, check.Max, check.Warning, check.Critical)
      } else {
        check.nagios.CheckThreshold(check.Name, sum, check.Warning, check.Critical)
      }
    }
  }
}

// checkPercentage checks the percentage of graphite targets against thresholds
func (check *Check) checkPercentage(data map[string]graphite.Target) {
  if len(data) % 2 == 0 {
    for i := 0; i < len(*check.TargetFlag); i += 2 {
      targetValue, targetValueOk := graphite.Find(data, (*check.TargetFlag)[i])
      targetValueOk = check.checkTargetValue(targetValue, targetValueOk, (*check.TargetFlag)[i])

      targetMax, targetMaxOk := graphite.Find(data, (*check.TargetFlag)[i+1])
      targetMaxOk = check.checkTargetValue(targetMax, targetMaxOk, (*check.TargetFlag)[i+1])

      if targetValueOk && targetMaxOk {
        check.checkPercentagePair(targetValue, targetMax)
      }
    }
  } else {
    check.nagios.Unknown("Count of targets is not even, can not calculate percentage for targets.")
  }
}

func (check *Check) checkTargetValue(target graphite.Target, ok bool, targetName string) bool {
  result := ok
  if ! ok {
    check.nagios.Unknown("target not found: "+targetName)
  }

  if target.IsEmpty() {
    result = false

    if check.EmptyOk {
      check.nagios.Ok("target is empty: "+target.Target())
    } else {
      check.nagios.Critical("target is empty: "+target.Target())
    }
  }
  return result
}

// checkPercentagePair checks the percentage of a pair of graphite targets against thresholds
func (check *Check) checkPercentagePair(value graphite.Target, max graphite.Target) {
  name := check.getName(value.Target())
  check.nagios.CheckPercentageThreshold(name, float64(value.Average()), float64(max.Average()), check.Warning, check.Critical)
}

// checkTargets checks targets against thresholds
func (check *Check) checkTargets(data map[string]graphite.Target) float64 {
  var sum float64
  for _, target := range *check.TargetFlag {
    if val, ok := data[target]; ok {
      if check.SumFlag {
        sum += float64(val.Average())
        check.Name = check.getName(target)
      } else {
        name := check.getName(target)
        check.nagios.CheckThreshold(name, float64(val.Average()), check.Warning, check.Critical)
      }
    } else {
      check.nagios.Unknown("target not found: "+target)
    }
  }
  return sum
}

func (check *Check) getName(target string) string {
  var name string
  if name = check.Name; check.Name == "" {
    if name = target; check.TargetNameRegexp != "" {
      pattern := regexp.MustCompile(check.TargetNameRegexp)
      if name = pattern.FindString(target); len(pattern.FindStringSubmatch(target)) > 1 {
        matches := pattern.FindStringSubmatch(target)
        name = strings.Join(matches[1:], ".")
      }
    }
  }
  return name
}
