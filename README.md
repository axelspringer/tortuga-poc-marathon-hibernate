# :zzz: Hiberthon

Hiberthon is a hibernation service for marathon

## ATTENTION This is a proof of concept. Do not run this in production

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