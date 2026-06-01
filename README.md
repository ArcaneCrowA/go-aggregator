# go-aggregator


Start server
```bash
docker compose up --build
```

Send request
```bash
curl localhost:8080/service/ -d '{"service_name":"yandex", "price":400, "user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","start_date":"07-2025"}'
```

Get 
```bash
# match all subscriptions for this user (no name filter)
curl "localhost:8080/service/?user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba"

# filter by name
curl "localhost:8080/service/?user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&service_name=yandex"

# filter by name + date range
curl "localhost:8080/service/?user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&service_name=yandex&start_date=07-2025&end_date=12-2025"

```
