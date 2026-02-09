# Transaction-indexer
## Overview
Tracks changes to desired addresses on remote bitcoin node.
## Endpoints
```
GET http://localhost:8080/healthcheck - check the state of the indexer

POST http://localhost:8080/register – register into indexer with basic auth (add "-U <username>:<password>" to your request)

POST http://localhost:8080/login - receive JWT token from indexer providing previously registered username and password (add "-U <username>:<password>" to your request)

POST http://localhost:8080/addresses - add new address to track. Needs previusly issued and still valid JWT. (add body "-d {addr:"<address>"}" to your request, add "-H "Authorization: Bearer <token>"" to your request)

GET http://localhost:8080/addresses - get the list of all tracked addresses by the registered user. (add "-H "Authorization: Bearer <token>"" to your request)

GET http://localhost:8080/addresses/<address>/balance - get the current (unmature) balance of the tracked address. Address should be added beforehand using POST /addresses method and JWT is needed.

POST http://localhost:8080/launch – launch the indexer. It would catch-up from the given block to the newest one and than make routine poll to synchronize with node every once in a while
```
## Run
1) Clone the repo:
```
git clone -b dev https://github.com/distraw/transaction-indexer.git
cd transaction-indexer
```
2) Run the tests:
```
go clean -testcache
go test ./e2e
```
Change the `config.local.yaml` to listen on desired node (leave as is if regtest is ok)
```
config.local.yaml

...
rpc:
  #host: bitcoin-testnet-rpc.publicnode.com:443
  host: rpcuser_0:rpcpassword_0@bitcoin:18443
  user: rpcuser_0
  pass: rpcpassword_0
  http_post_mode: true
  disable_tls: true
...
```
3) Build docker image (not necessary if you ran e2e tests, as they already built it for you)
```
docker build -t transaction-indexer .
```
4) Start the compose
```
docker compose down -v
docker compose up
```
5) Register into indexer
```
curl -X GET http://localhost:8080/healthcheck
curl -X POST -U user:123 http://localhost:8080/register
curl -X POST -U user:123 http://localhost:8080/login
```
6) Using returned JWT token, add desired addresses to track
```
curl -X POST \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer <token>"
     -d '{"addr":"<address>"}' \
     http://localhost:8080/addresses
```
7) After everything is setupped, launch the indexer to do initial catch-up and poll routinely on the node
```
curl -X POST http://localhost:8080/launch
```
8) Track your balances
```
curl -X POST \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer <token>"
     http://localhost:8080/addresses/<address>/balance
```