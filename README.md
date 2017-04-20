# AppCop [![Build](https://travis-ci.org/allegro/marathon-appcop.svg?branch=master)](https://travis-ci.org/allegro/marathon-appcop.svg?branch=master) [![Coverage Status](https://coveralls.io/repos/github/allegro/marathon-appcop/badge.svg?branch=master)](https://coveralls.io/github/allegro/marathon-appcop?branch=master)

`Marathon AppCop` - Marathon applications law enforcement.

In large [Mesos](mesos.apache.org) deployments there could be thousands of applications running and deploying every day.
Sometimes they happen to be broken, forgotten and unmaintained which could exert pressure on cluster in numerous ways.

To address that AppCop clears [Marathon](https://github.com/mesosphere/marathon) from broken application deployments.

## How it works

`AppCop` takes information provided by the [Marathon event-stream](https://mesosphere.github.io/marathon/docs/event-bus.html)
related to applications failures and scales them down.

### Scoring Mechanism

Based on Marathon events (TASK_KILL, TASK_FAIL, TASK_FINISHED),
AppCop is building score registry for each application event emited.
Each score is incremented by each app event, so if events related to failures are comming it
is constantly raising.
When application passes treshold, then AppCop scales application one instance down forcefully and put appcop label in app definition. After that, score for this application is reset.
When there is only one instance, then and score is pass theshold then application is suspended.
Scores are periodically reset.

### GarbageCollection

AppCop is periodically fetching applications and groups from Marathon.
When application is suspended or group is empty for long (configurable) time then it is deleted.


## Installation

### Installing from source code

To simply compile and run the source code:

```
go run main.go [options]
```

To run the tests:

```
make test
```

To build the binary:

```
make build
```

To build deb package:
```
make pack
```

Check dist/ dir.

## Setting up `AppCop`

AppCcop should be installed on all Marathon masters.
The event subscription should be set to `localhost` to reduce network traffic.
Please refer to options section for more.


## `Marathon Labels`

`AppCop` is using `Marathon` labels to communicate actions or to tune execution logic.

Used labels:

Name                      |       Possible values     |    r/w   |    Description
--------------------------|---------------------------|----------|------------------
appcop                    | `suspend`, `scaleDown`    |    w     | Every time `AppCop` scales or suspend application, put appropriate label in app definition
APP_IMMUNITY              |   `false`, `true`         |    r     | When AppCop encounters this label in app definition, treats it as immune to all penalties (excused from all criminal acts on cluster). Use this feature wisely, because if applied to often it could defeat whole purpose for using AppCop

r - label is taken from app definition, not altered,
w - label is manipulated by `AppCop`.

### Options

Argument                    | Default           | Description
----------------------------|-------------------|------------------------------------------------------
config-file                 |                   | Path to a JSON file to read configuration from. Note: Will override options set earlier on the command line
event-stream-location       | /v2/events        | Get events from this stream
my-leader                   | marathon-dev      | My leader, when Marathon /v2/leader endpoint return the same string as this one, make subscription to event stream and launch jobs.
events-queue-size           | `1000`            | Size of events queue
listen                      | `:4444`           | Accept connections at this address
log-file                    |                   | Save logs to file (e.g.: `/var/log/appcop.log`). If empty logs are published to STDERR
log-format                  | `text`            | Log format: JSON, text
log-level                   | `info`            | Log level: panic, fatal, error, warn, info or debug
marathon-location           | `example.com:8080`| Marathon URL
marathon-password           |                   | Marathon password for basic auth
marathon-protocol           | `http`            | Marathon protocol (http or https)
marathon-ssl-verify         | `true`            | Verify certificates when connecting via SSL
marathon-timeout            | `30s`             | Time limit for requests made by the Marathon HTTP client. A timeout of zero means no timeout
marathon-username           |                   | Marathon username for basic auth
scale-down-score            | `30`              | Score for application to scale it one instance down
scale-limit                 | `2`               | How many scale down actions to commit in one scaling down iteration
update-interval             | `2s`              | Interval for updating app scores
reset-interval              | `1d`              | How often collected scores are reset
evaluate-interval           | `30s`             | How often collected scores are compared against scale-down-score
metrics-interval            | `30s`             | Metrics reporting interval
metrics-location            |                   | Graphite URL (used when metrics-target is set to graphite)
metrics-prefix              | `default`         | Metrics prefix (default is resolved to <hostname>.<app_name>
metrics-target              | `stdout`          | Metrics destination stdout or graphite (empty string disables metrics)
workers-pool-size           | `10`              | Number of concurrent workers processing events
mgc-enabled                 | `true`            | Enable garbage collecting of Marathon, old suspended applications will be deleted
mgc-max-suspend-time        | `7 days`          | How long application should be suspended before deleting it
mgc-interval                | `8 hours`         | Marathon GC interval
mgc-appcop-only             | `true`            | Delete only applications suspended by AppCop
dry-run                     | `false`           | Perform a trial run with no changes made to marathon


### Endpoints

Endpoint  | Description
----------|------------------------------------------------------------------------------------
`/health` | healthcheck - returns `OK`
