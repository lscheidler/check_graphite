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

// Package main contains check_graphite
package main

import (
  "flag"
  "fmt"
  "math"
  "regexp"
  "os"
  "strings"
  "time"

  "github.com/lscheidler/go-nagios"
  "github.com/lscheidler/check_graphite/graphite"
)

const Version = "0.1.3"

// flag variables
var (
  critical float64
  emptyOk bool
  from string
  graphiteUrl string
  name string
  max float64
  percentage bool
  perfdata bool
  sumFlag bool
  targetNameRegexp string
  timeout time.Duration
  until string
  warning float64
)

type target []string

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (i *target) String() string {
  return fmt.Sprint(*i)
}

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
// It's a comma-separated list, so we split it.
func (i *target) Set(value string) error {
  for _, t := range strings.Split(value, ",") {
    *i = append(*i, t)
  }
  return nil
}

var targetFlag target

// init parses command line arguments
func init() {
  flag.Usage = func() {
    fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s (%s):\n", os.Args[0], Version)
    flag.PrintDefaults()

    examples := `
examples:
  # check target againts thresholds
  ` + os.Args[0] + ` -u http://localhost:8080 -target=collectd.host1.memory.percent-used -c 90 -w 80

  # calculate percentage in pairs of targets and check percentage againts thresholds
  ` + os.Args[0] + ` -u http://localhost:8080 -target=collectd.host1.GenericJMX-app1_memory-heap.memory-{used,max} -target=collectd.host1.GenericJMX-app1_memory_pool-CMS_Perm_Gen.memory-{used,max} -p -c 42 -w 2 -r "memory[^-]*-(.*)$"

  # check sum of targets against thresholds
  ` + os.Args[0] + ` -u http://localhost:8080 -target=collectd.host1.aggregation-cpu-average.cpu-{system,user} -w 70 -c 90 -s -name cpu-usage`

    fmt.Fprintf(flag.CommandLine.Output(), "%s\n", examples)
  }

  const (
    criticalUsage               = "set critical threshold"
    emptyOkUsage                = "empty data from graphite is ok"
    emptyOkDefaultVal           = false
    fromDefaultVal              = "-2min"
    fromUsage                   = "set from"
    graphiteUrlDefaultVal       = ""
    graphiteUrlUsage            = "graphite url"
    nameDefaultVal              = ""
    nameUsage                   = "set name to use in message and perfdata for target"
    maxUsage                    = "Maximum value for target"
    percentageDefaultVal        = false
    percentageUsage             = "calculate percentage and check against thresholds, uses -max or every second target as max for previous target"
    perfdataDefaultVal          = false
    perfdataUsage               = "show perfdata"
    sumDefaultVal               = false
    sumUsage                    = "summarize target datapoints and check against thresholds"
    targetDefaultVal            = ""
    targetUsage                 = "comma-separated list of targets to check"
    targetNameRegexpDefaultVal  = ""
    targetNameRegexpUsage       = "regex to show only the part of the target name, which matches"
    untilDefaultVal             = "now"
    untilUsage                  = "set until"
    warningUsage                = "set warning threshold"
  )

  flag.Float64Var(&critical,          "critical",             math.NaN(), criticalUsage)
  flag.Float64Var(&critical,          "c",                    math.NaN(), criticalUsage)
  flag.BoolVar(&emptyOk,              "empty-ok",             emptyOkDefaultVal, emptyOkUsage)
  flag.StringVar(&from,               "from",                 fromDefaultVal, fromUsage)
  flag.StringVar(&from,               "f",                    fromDefaultVal, fromUsage)
  flag.StringVar(&graphiteUrl,        "graphite-url",         graphiteUrlDefaultVal, graphiteUrlUsage)
  flag.StringVar(&graphiteUrl,        "u",                    graphiteUrlDefaultVal, graphiteUrlUsage)
  flag.StringVar(&name,               "name",                 nameDefaultVal, nameUsage)
  flag.Float64Var(&max,               "max",                  math.NaN(), maxUsage)
  flag.Float64Var(&max,               "m",                    math.NaN(), maxUsage)
  flag.BoolVar(&percentage,           "percentage",           percentageDefaultVal, percentageUsage)
  flag.BoolVar(&percentage,           "p",                    percentageDefaultVal, percentageUsage)
  flag.BoolVar(&perfdata,             "perfdata",             perfdataDefaultVal, perfdataUsage)
  flag.BoolVar(&sumFlag,              "summarize",            sumDefaultVal, sumUsage)
  flag.BoolVar(&sumFlag,              "s",                    sumDefaultVal, sumUsage)
  flag.Var(&targetFlag,               "target",               targetUsage)
  flag.Var(&targetFlag,               "t",                    targetUsage)
  flag.StringVar(&targetNameRegexp,   "target-name-regexp",   targetNameRegexpDefaultVal, targetNameRegexpUsage)
  flag.StringVar(&targetNameRegexp,   "r",                    targetNameRegexpDefaultVal, targetNameRegexpUsage)
  flag.DurationVar(&timeout,          "timeout",              10, "timeout")
  flag.StringVar(&until,              "until",                untilDefaultVal, untilUsage)
  flag.Float64Var(&warning,           "warning",              math.NaN(), warningUsage)
  flag.Float64Var(&warning,           "w",                    math.NaN(), warningUsage)
}

// check checks graphite targets against thresholds
func check(nagios *nagios.Nagios, data map[string]graphite.Target) {
  if percentage && math.IsNaN(max) {
    checkPercentage(nagios, data)
  } else {
    sum := checkTargets(nagios, data)

    if sumFlag {
      if percentage && ! math.IsNaN(max) {
        nagios.CheckPercentageThreshold(name, sum, max, warning, critical)
      } else {
        nagios.CheckThreshold(name, sum, warning, critical)
      }
    }
  }
}

// checkPercentage checks the percentage of graphite targets against thresholds
func checkPercentage(nagios *nagios.Nagios, data map[string]graphite.Target) {
  if len(data) % 2 == 0 {
    useTargetName := name == ""
    for i := 0; i < len(targetFlag); i += 2 {
      targetValue, targetValueOk := graphite.Find(data, targetFlag[i])
      targetValueOk = checkTargetValue(nagios, targetValue, targetValueOk, targetFlag[i])

      targetMax, targetMaxOk := graphite.Find(data, targetFlag[i+1])
      targetMaxOk = checkTargetValue(nagios, targetMax, targetMaxOk, targetFlag[i+1])

      if targetValueOk && targetMaxOk {
        checkPercentagePair(nagios, targetValue, targetMax, useTargetName)
      }
    }
  } else {
    nagios.Unknown("Count of targets is not even, can not calculate percentage for targets.")
  }
}

func checkTargetValue(nagios *nagios.Nagios, target graphite.Target, ok bool, targetName string) bool {
  result := ok
  if ! ok {
    nagios.Unknown("target not found: "+targetName)
  }

  if target.IsEmpty() {
    result = false

    if emptyOk {
      nagios.Ok("target is empty: "+target.Target())
    } else {
      nagios.Critical("target is empty: "+target.Target())
    }
  }
  return result
}

// checkPercentagePair checks the percentage of a pair of graphite targets against thresholds
func checkPercentagePair(nagios *nagios.Nagios, value graphite.Target, max graphite.Target, useTargetName bool) {
  if useTargetName {
    if name = value.Target(); targetNameRegexp != "" {
      pattern := regexp.MustCompile(targetNameRegexp)
      if name = value.Target(); pattern.MatchString(value.Target()) {
        if name = pattern.FindString(value.Target()); len(pattern.FindStringSubmatch(value.Target())) > 1 {
          matches := pattern.FindStringSubmatch(value.Target())
          name = strings.Join(matches[1:], ".")
        }
      }
    }
  }
  nagios.CheckPercentageThreshold(name, float64(value.Average()), float64(max.Average()), warning, critical)
}

// checkTargets checks targets against thresholds
func checkTargets(nagios *nagios.Nagios, data map[string]graphite.Target) float64 {
  var sum float64
  for _, target := range targetFlag {
    if val, ok := data[target]; ok {
      if sumFlag {
        sum += float64(val.Average())
        if name == "" {
          if targetNameRegexp == "" {
            name = target
          } else {
            pattern := regexp.MustCompile(targetNameRegexp)
            name = string(pattern.Find([]byte(target)))
          }
        }
      } else {
        nagios.CheckThreshold(target, float64(val.Average()), warning, critical)
      }
    } else {
      nagios.Unknown("target not found: "+target)
    }
  }
  return sum
}

// main initialize nagios struct and run check
func main() {
  flag.Parse()

  nagios := nagios.Init()
  nagios.ShowPerfdata = perfdata
  defer nagios.Exit()

  data, err := graphite.GetTargets(graphiteUrl, targetFlag, from, until, timeout)
  if err != nil {
    nagios.Unknown(err.Error())
    return
  }

  check(&nagios, data)
}
