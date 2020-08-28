/*
Copyright 2020 Lars Eric Scheidler

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

import (
	"flag"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/lscheidler/check_graphite/graphite"
	"github.com/lscheidler/go-nagios"
)

// flag variables
var (
	from        string
	graphiteURL string
	targetFlag  target
	timeout     time.Duration
	until       string
)

// Init initializes command line interface
func Init(version string) *Check {
	checkGraphite := Check{
		nagios: nagios.Init(),
	}

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s (%s):\n", os.Args[0], version)
		flag.PrintDefaults()

		examples := `
examples:
  # check target against thresholds
  ` + os.Args[0] + ` -u http://localhost:8080 -target=collectd.host1.memory.percent-used -c 90 -w 80

  # calculate percentage in pairs of targets and check percentage against thresholds
  ` + os.Args[0] + ` -u http://localhost:8080 -target=collectd.host1.GenericJMX-app1_memory-heap.memory-{used,max} -target=collectd.host1.GenericJMX-app1_memory_pool-CMS_Perm_Gen.memory-{used,max} -p -c 42 -w 2 -r "memory[^-]*-(.*)$"

  # check sum of targets against thresholds
  ` + os.Args[0] + ` -u http://localhost:8080 -target=collectd.host1.aggregation-cpu-average.cpu-{system,user} -w 70 -c 90 -s -name cpu-usage`

		fmt.Fprintf(flag.CommandLine.Output(), "%s\n", examples)
	}

	const (
		criticalUsage              = "set critical threshold"
		debugUsage                 = "set debug mode"
		debugDefaultVal            = false
		emptyOkUsage               = "empty data from graphite is ok"
		emptyOkDefaultVal          = false
		fromDefaultVal             = "-2min"
		fromUsage                  = "set from"
		graphiteURLDefaultVal      = ""
		graphiteURLUsage           = "graphite url"
		nameDefaultVal             = ""
		nameUsage                  = "set name to use in message and perfdata for target"
		maxUsage                   = "Maximum value for target"
		percentageDefaultVal       = false
		percentageUsage            = "calculate percentage and check against thresholds, uses -max or every second target as max for previous target"
		perfdataDefaultVal         = false
		perfdataUsage              = "show perfdata"
		sumDefaultVal              = false
		sumUsage                   = "summarize target datapoints and check against thresholds"
		targetDefaultVal           = ""
		targetUsage                = "comma-separated list of targets to check"
		targetNameRegexpDefaultVal = ""
		targetNameRegexpUsage      = "regex to show only the part of the target name, which matches"
		untilDefaultVal            = "now"
		untilUsage                 = "set until"
		warningUsage               = "set warning threshold"
	)

	flag.Float64Var(&checkGraphite.Critical, "critical", math.NaN(), criticalUsage)
	flag.Float64Var(&checkGraphite.Critical, "c", math.NaN(), criticalUsage)
	flag.BoolVar(&checkGraphite.Debug, "d", debugDefaultVal, debugUsage)
	flag.BoolVar(&checkGraphite.EmptyOk, "empty-ok", emptyOkDefaultVal, emptyOkUsage)
	flag.StringVar(&from, "from", fromDefaultVal, fromUsage)
	flag.StringVar(&from, "f", fromDefaultVal, fromUsage)
	flag.StringVar(&graphiteURL, "graphite-url", graphiteURLDefaultVal, graphiteURLUsage)
	flag.StringVar(&graphiteURL, "u", graphiteURLDefaultVal, graphiteURLUsage)
	flag.StringVar(&checkGraphite.Name, "name", nameDefaultVal, nameUsage)
	flag.Float64Var(&checkGraphite.Max, "max", math.NaN(), maxUsage)
	flag.Float64Var(&checkGraphite.Max, "m", math.NaN(), maxUsage)
	flag.BoolVar(&checkGraphite.Percentage, "percentage", percentageDefaultVal, percentageUsage)
	flag.BoolVar(&checkGraphite.Percentage, "p", percentageDefaultVal, percentageUsage)
	flag.BoolVar(&checkGraphite.Perfdata, "perfdata", perfdataDefaultVal, perfdataUsage)
	flag.BoolVar(&checkGraphite.SumFlag, "summarize", sumDefaultVal, sumUsage)
	flag.BoolVar(&checkGraphite.SumFlag, "s", sumDefaultVal, sumUsage)
	flag.Var(&targetFlag, "target", targetUsage)
	flag.Var(&targetFlag, "t", targetUsage)
	flag.StringVar(&checkGraphite.TargetNameRegexp, "target-name-regexp", targetNameRegexpDefaultVal, targetNameRegexpUsage)
	flag.StringVar(&checkGraphite.TargetNameRegexp, "r", targetNameRegexpDefaultVal, targetNameRegexpUsage)
	flag.DurationVar(&timeout, "timeout", 10, "timeout")
	flag.StringVar(&until, "until", untilDefaultVal, untilUsage)
	flag.Float64Var(&checkGraphite.Warning, "warning", math.NaN(), warningUsage)
	flag.Float64Var(&checkGraphite.Warning, "w", math.NaN(), warningUsage)

	return &checkGraphite
}

// Run parses command line arguments and run check
func (check *Check) Run() {
	flag.Parse()
	check.TargetFlag = &targetFlag

	defer check.Exit()

	data, err := graphite.GetTargets(graphiteURL, targetFlag, from, until, timeout)
	if err != nil {
		check.Error(err)
		return
	}

	check.runCheck(data)
}
