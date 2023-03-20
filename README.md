# fluentbit-out-pgsql

### build as .so
```shell
go build -buildmode=c-shared -o out_pgsql.so main.go
```

### ldd the .so
```shell
ldd out_pgsql.so
```

### define plugins files in main configuration file
```shell
[SERVICE]
    plugins_file /fbtest/plugins.conf

[INPUT]
    Name dummy

[OUTPUT]
    Name stdout
```
### content of plugin file
```shell
[PLUGINS]
    Path /fbtest/out_pgsql.so
```
### apply new output plugin in main configuration file
```shell
[SERVICE]
    Flush        5
    Daemon       Off
    Log_Level    info
    plugins_file /fbtest/plugins.conf

[INPUT]
    Name  cpu
    tag   cpu.local

[OUTPUT]
    name            stdout
    match           *

[OUTPUT]
    name            out_pgsql
    pghost          127.0.0.1
    pgport          5432
    user            postgres
    password        xxx
    match           *
    database        xxx
    table           xxx
```


