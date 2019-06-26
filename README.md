# Superman Detector

Application that identifies logins by a user that occur from
locations that are farther apart than a normal person can reasonably travel on an airplane.
These locations were determined by looking up the source IP addresses of successful logins in
a GeoIP database. This behavior is interesting because it can be a strong indication that a user
has had their account credentials compromised.

## Notes
Project Submission

## Getting Started
``git clone git@github.com:edwardsb/secureworks.git``

### Prerequisites
- Docker (for build)
- Golang 1.11+ (for code)

### Installing
`make build` - for local install

`docker build .` - for docker

`docker-compose up -d` - to get the system up and running

## Running the tests
`go test`


## Usage
```
curl -X POST \
  http://localhost:3000/v1/ \
  -H 'Content-Type: application/json' \
  -H 'cache-control: no-cache' \
  -d '{
	"username": "user2",
    "unix_timestamp": 1561600005,
    "event_uuid": "05d86fca-825e-4515-86cc-7775a2d8047e",
    "ip_address": "68.193.88.103"
}'
```


## Dependencies
- [Sqlite](https://www.sqlite.org/index.html) - Database
- [Sqlite3](https://github.com/mattn/go-sqlite3) - Go Sqlite Driver
- [Sqlx](https://github.com/jmoiron/sqlx) - Extensions to the standard sql lib
- [Chi](https://github.com/go-chi/chi) - Router
- [GeoIP](https://github.com/oschwald/geoip2-golang) - GeoIP Data
- [MultiError](https://github.com/hashicorp/go-multierror) - Collect Many Errors from Schema Validation
- [Sql Mock](https://github.com/DATA-DOG/go-sqlmock) - SQL Mocking
- [Cobra](https://github.com/spf13/cobra) - CLI
- [Viper](https://github.com/spf13/viper) - Configuration
- [Haversine](https://github.com/umahmood/haversine) - Distance on a sphere formula lib
- [JSONSchema](https://github.com/xeipuuv/gojsonschema) - JSONSchema Validation
- [Testify](https://github.com/stretchr/testify) - Testing helpers

##  Authors <a name = "authors"></a>
- [Ben Edwards](https://github.com/edwardsb) - Project Submitter
