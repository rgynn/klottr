# klottr backend

Service in the same vein as hackernews or reddit. Threads, comments, users, upvotes. Content posted by users to site only available for a week.

## Required .env file to run locally

```
DEBUG=true
HOST=0.0.0.0
PORT=3000
REQBODYLIMIT_BYTES=1000000
TIMEOUT_REQ=6s
TIMEOUT_IDLE=3s
TIMEOUT_READ=3s
TIMEOUT_WRITE=3s
CORS_ALLOW_ORIGINS=localhost
POST_TTL_SECONDS=86400
DATABASE_URL=mongodb+srv://<username>:<password>@<hostname>/<defaultdb>?authSource=admin&replicaSet=<replicasetname>&tls=true&tlsCAFile=<filepath>
DATABASE_NAME=***
JWT_SECRET=***
```

## Prerequisites
* MongoDB database provisioned, and .env file DATABASE_URL and DATABASE_NAME connection string filled in correctly

## How to run locally
1. Make sure .env file is present and filled in correctly
2. Run ``make db_seed`` to setup collections and indexes
3. Run ``make test_int`` to make sure all intergration test run OK
4. Run ``make db_seed`` again to reset database collections
5. Run command: ``make run`` or ``go run main.go`` to run the service locally

## Docker
Build a docker image with the ``make build_docker`` target.
Make sure the environment variables are all provided when running the container.

## Testing
Unit testing easily perfomed by either using ``go test./...`` or ``make test``
Run integration tests using ``make test_intg``