# :zzz: Hiberthon

Hiberthon is a hibernation service for marathon

## ATTENTION This is a proof of concept. Do not run this in production

## Build

Build hiberthon

    make generate
    make build/hiberthon/static

Build the trigger

    make build/hiberthon/static

## Tag the tasks in marathon

Set ```"hiberthon.enable": "true"``` to activate hibernation for this application. Set ```"hiberthon.group": "0CD7EB62-1ACF-4EB3-B776-8CBF190D410D"``` to the id.

    {
        "labels": {
            "hiberthon.enable": "true",
            "hiberthon.group": "0CD7EB62-1ACF-4EB3-B776-8CBF190D410D",
            "traefik.enable": "true"
        }
    }

## Create DynamoDB table

Please change the RCU and WCU to your own needs

    aws dynamodb create-table --table-name Hibernate --attribute-definitions AttributeName=id,AttributeType=S --key-schema AttributeName=id,KeyType=HASH --provisioned-throughput ReadCapacityUnits=1,WriteCapacityUnits=1

## Database host records

    {
        "id": "0CD7EB62-1ACF-4EB3-B776-8CBF190D410D",
        "hosts": [
            "backend.testing.example.org",
            "frontend.testing.example.org"
        ],
        "state": "run",
        "latestUsage": 1543941897,
        "actionNotBefore": 1543941897,
        "scaleMap": {
            "/testing/memcache-test/memcache": 1,
            "/testing/frontend-test/frontend": 1,
            "/testing/backend-test/backend": 1
        },
        "idleDuration": 120
    }

## Parameter

* -db-endpoint DynamoDB endpoint (Required)
* -db-region DynamoDB region (Required)
* -db-key DynamoDB credential key
* -db-secret DynamoDB credential secret
* -marathon-endpoint DynamoDB endpoint (Required)
* -listener Web listener (Required)

## Environment variables

Every parameter is also available as env var. E.g. -db-endpoint becomes HIBERTHON_DB_ENDPOINT etc. pp.

## :punch: Hiberthon trigger

Hiberthon trigger is a little helper tool to tell if a host is active. This can be done by read access logs along.

### Trigger Parameter

* -endpoint *url* Endpoint for the hiberthon api (Required)
* -logfile *path* Path to the logfile to watch (Required)
* -format *fmt* "traefik:clf" is currently the only one (Required)
* -collection-time *time* in seconds (default 10s)
* -host-update-time *time* in seconds (default 30s)

### Example

Assume a access log format like traefiks CLF

    10.255.0.14 - - [12/Mar/2018:16:57:53 +0000] "POST /api/v4/jobs/request HTTP/1.1" 204 0 - "gitlab-runner 10.0.2 (10-0-stable; go1.8.3; windows/amd64)" 147 "Host-gitlab-14" "http://10.0.0.19:80" 4ms

    ./bin/hiberthon-trigger -format "traefik:clf" -logfile traefik.log -endpoint http://hiberthon/-/api/trigger -collection-time 5 -host-update-time 10