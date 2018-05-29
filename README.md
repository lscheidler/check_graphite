# Check\_Graphite

## Usage

### check target againts thresholds
```
./check_graphite -u http://localhost:8080 -target=collectd.host1.memory.percent-used -c 90 -w 80
```

### calculate percentage in pairs of targets and check percentage againts thresholds
```
./check_graphite -u http://localhost:8080 -target=collectd.host1.GenericJMX-app1_memory-heap.memory-{used,max} -target=collectd.host1.GenericJMX-app1_memory_pool-CMS_Perm_Gen.memory-{used,max} -p -c 42 -w 2 -r "memory[^-]*-(.*)$"
```

### check sum of targets against thresholds
```
./check_graphite -u http://localhost:8080 -target=collectd.host1.aggregation-cpu-average.cpu-{system,user} -w 70 -c 90 -s -name cpu-usage
```

## Contributing

Bug reports and pull requests are welcome on GitHub at https://github.com/lscheidler/check_graphite.

## License

The code is available as open source under the terms of the [Apache 2.0 License](http://opensource.org/licenses/Apache-2.0).
