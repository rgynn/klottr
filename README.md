# klottr backend

Service in the same vein as hackernews. But content posted by users to site only available for 24h.

## Required .env file to run locally

```
DEBUG=true
HOST=0.0.0.0
PORT=3000
REQBODYLIMIT=100K
TIMEOUT_REQ=6s
TIMEOUT_IDLE=3s
TIMEOUT_READ=3s
TIMEOUT_WRITE=3s
CORS_ALLOW_ORIGINS=localhost
POST_TTL=24h
DATABASE_URL=mongodb+srv://<username>:<password>@<hostname>/<defaultdb>?authSource=admin&replicaSet=<replicasetname>&tls=true&tlsCAFile=<filepath>
DATABASE_NAME=***
JWT_SECRET=***
```

## How to run
1. Make sure .env file is present
2. Run command: ``make run`` or ``go run main.go``