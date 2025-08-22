## CodeXray


## Build agent
swag init
go mod tidy
go build .


## run agent on port 8080
go run .

## send skywalking data on otel collector and otel format
curl -X 'POST' \
  'http://localhost:8081/send-to-otel' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "traceId": "abc123",
  "service": "my-dotnet-service5",
  "serviceInstance": "instance-uuid-01",
  "spans": [
    {
      "spanId": 0,
      "parentSpanId": -1,
      "operationName": "/api/users",
      "spanType": "Entry",
      "isError": 0,
      "startTime": 1750780200000,
      "endTime": 1750780203000,
      "peer": "",
      "component": "HTTP",
      "layer": "HTTP",
      "tags": [
        { "key": "http.method", "value": "GET" },
        { "key": "http.status_code", "value": "200" }
      ]
    },
    {
      "spanId": 1,
      "parentSpanId": 0,
      "operationName": "Call MySQL",
      "spanType": "Exit",
      "isError": 0,
      "startTime": 1750780201000,
      "endTime": 1750780202000,
      "peer": "mysql-server:3306",
      "component": "MySQL",
      "layer": "Database",
      "tags": [
        { "key": "db.type", "value": "mysql" },
        { "key": "db.statement", "value": "SELECT * FROM users" }
      ]
    },
    {
      "spanId": 2,
      "parentSpanId": 1,
      "operationName": "Redis Cache",
      "spanType": "Local",
      "isError": 1,
      "startTime": 1750780202000,
      "endTime": 1750780203000,
      "peer": "redis-server:6379",
      "component": "Redis",
      "layer": "Cache",
      "tags": [
        { "key": "cache.hit", "value": "false" }
      ]
    }
  ]
}
'

## Test in local (swagger) (skywalking payload in skywalking/skywalking.json)
http://localhost:8080/swagger/index.html

## test otel payload direct on collector (check timestamp first in trace.json)
curl.exe -v -X POST http://labs.codexray.io:8000/v1/traces `
   -H "Content-Type: application/json" `
   --data-binary "@trace.json"


