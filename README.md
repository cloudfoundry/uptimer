# Uptimer
Uptimer is a tool for measuring CF availability
during an arbitrary operation of interest.

It measures:
- app availability,
  by making frequent http requests.
- logging availability,
  by periodically fetching recent logs,
  and by periodically initiating live log streaming.
- push availability,
  by periodically pushing a very simple app.
- app syslog availability,
  by periodically checking that app logs
  drain to a syslog sink.
- app stats availability,
  by periodically checking that app stats
  are not unavailable.

It is often used to monitor availability
during upgrade deployments.

## Installation

```
go get github.com/cloudfoundry/uptimer
```

## Usage
`uptimer -configFile config.json [-resultFile result.json]`.

Uptimer needs configuration to run.
It reads a `json` file
specified with the `-configFile` flag.
The config describes how to target the CF you wish to measure,
and commands to run while measuring uptime.

Uptimer can optionally be given a resultFile (`-resultFile) to which 
resultant measurements will be written in json format.

## Config
Here is an example config `json`:
```
{
    "while": [
        {
            "command": "bosh",
            "command_args": ["deploy", "-d", "cf", "interpolated-cf-deployment.yml", "-n"]
        },
        {
            "command": "bosh",
            "command_args": ["delete-deployment", "-d", "cf", "-n"]
        }
    ],
    "cf": {
        "api": "api.my-cf.com",
        "app_domain": "my-cf.com",
        "admin_user": "admin",
        "admin_password": "PASS",
        "tcp_domain": "tcp.my-cf.com",
        "use_single_app_instance": false,
        "available_port": 1025
    },
    "optional_tests": {
      "run_app_syslog_availability": true
    },
    "allowed_failures": {
        "app_pushability": 2,
        "http_availability": 5,
        "app_stats": 2,
        "recent_logs": 2,
        "streaming_logs": 2,
        "app_syslog_availability": 2
    }
}
```
### While (required)
The `while` section is an array of commands.
These are executed in order while the measurement is run.
When the last while command exits,
uptimer will conclude its measurements
and print a summary of results.
If a `while` command exits with a non-zero exit code,
uptimer won't run subsequent `while` commands.
It will conclude its measurements and report,
then exit with exit code 64.

A single-command `while` array is fine,
and you can use sleep
to just run up-timer for some period:
```
"while": [{
    "command": "sleep",
    "command_args": ["600"]
}]
```

### Cf (mostly required)
The `cf` section contains information necessary
to perform the `cf auth` and `cf login` commands
on the target environment.

Uptimer requires an admin user
because it creates and configures an org and space
during test setup.

The `tcp_domain` and `available_port` values
are not required
_unless_ you elect to run the `app_syslog_availability` test.

Uptimer by default pushes two instances of an app
for its uptime measurements,
but it can be configured to push only a single instance
by setting the optional
`use_single_app_instance` value to
`true` under the `cf` section.

Only set this flag to `true`
if you understand
the implications it has
for your uptime measurements.

### Creating TCP Domain (optional)
If running `run_tcp_availability` or `run_app_syslog_availability`
optional tests, you must create a tcp domain on your environment prior
to running uptimer. To create a TCP domain, run:
```
cf create-shared-domain tcp.[SYSTEM_DOMAIN] --router-group default-tcp
```
Save the value of `tcp.[SYSTEM_DOMAIN]` in the uptimer config under `tcp_domain`.

### Optional tests (optional)
The `optional_tests` section is optional,
as are each entry in the section.
If these values are omitted,
they are assumed to be false.

For the `run_tcp_availability` test,
TCP routing is required,
and you must specify
the `tcp_domain` and `tcp_port` values
in the `Cf` section of the configuration.

For the `run_app_syslog_availability` test,
TCP routing is required,
and you must specify
the `tcp_domain` and `available_port` values
in the `Cf` section of the configuration.

### Allowed Failures (optional)
The `allowed_failures` section contains failure thresholds,
expressed as integers.
In the example above, the `app_pushability` measurement
can fail at most two times before
the measurement is marked as failed.

This allows users to tolerate variance in downtime
to suit their needs.

If this config section is not specified,
the default threshold will be 0
for each measurement.

## CI
If you wish to run uptimer in CI
during bosh deployments specifically,
the deploy tasks in [cf-deployment-concourse-tasks](https://github.com/cloudfoundry/cf-deployment-concourse-tasks)
explicitly support this as an option.
If you wish to use it some other way,
said tasks may nonetheless be a useful example.
