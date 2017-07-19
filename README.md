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

The CF Release Integration team uses it
to monitor availability during migrations
from `cf-release` to `cf-deployment`,
and during upgrade deployments.

## Usage
`uptimer -configFile config.json`.

Uptimer needs configuration to run.
It reads a `json` file
specified with the `-configFile` flag.
The config describes how to target the CF you wish to measure,
and commands to run while measuring uptime.

## Config
Here is an example config `json`:
```
{
    "while": [
        {
            "command": "bosh",
            "command_args": ["deploy"]
        },
        {
            "command": "bosh",
            "command_args": ["delete-deployment"]
        }
    ],
    "cf": {
        "api": "api.my-cf.com",
        "app_domain": "my-cf.com",
        "admin_user": "admin",
        "admin_password": "PASS"
    }
}
```
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

The `cf` section contains information necessary
to perform the `cf auth` and `cf login` commands
on the target environment.

Uptimer requires an admin user
because it creates and configures an org and space
during test setup.

## CI
If you wish to run uptimer in CI
during bosh deployments specifically,
the deploy tasks in [cf-deployment-concourse-tasks](https://github.com/cloudfoundry/cf-deployment-concourse-tasks)
explicitly support this as an option.
If you wish to use it some other way,
said tasks may nonetheless be a useful example.
