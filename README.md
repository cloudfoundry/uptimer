# Uptimer is a tool for monitoring CF
Uptimer is a tool for measuring availability of an app.
Eventually, it will also test
CF control-plane functionality.

# CF Operators and developers may find it useful
This tool is for CF operators
to monitor performance characteristics
of their CF instance
during some operation.
For example, the release integration team uses it
to monitor uptime during migrations
from CF Release to CF Deployment.

# To use uptimer, configure it with an operation to monitor

Create a `config.json` file
that tells `uptimer`
how to target
the CF you wish to test
and a command (e.g. `bosh deploy`)
to run while testing uptime.

## You will need to configure it with a `while_command`
A `while_command` is a command
that is executed
to determine
when to stop running the uptime tests.
When the while command finishes,
uptimer will stop measuring uptime
and print a summary of results.
For example,
you may wish to configure
the while_command to be
a `bosh deploy` command.

## You will also need to configure it with credentials and an app to use
You should pass the
`api` endpoint,
`apps_domain`,
`admin` credentials,
and the path to an app to push.
It requires an admin user because
it will attempt
to create an org and space
and push the configured app to that space. 
Here is an example `config.json`:
```
{
    "while": {
        "command": "sleep",
        "command_args": ["30"]
    },
    "cf": {
        "api": "api.my-cf.com",
        "app_domain": "my-cf.com",
        "admin_user": "admin",
        "admin_password": "PASS",
        "app_path": "/Users/somebody/workspace/my_app"
    }
}
```

# To use this in CI, you will need to create your own concourse task
We have not implemented
a concourse task
to run uptimer ourselves yet,
and we are not sure yet
if such a task would useful enough to publish.
