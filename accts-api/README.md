Spin up mongo cluster
```
./setup.sh
```

Test Apis Locally

```sh 
# Register a new account from CLI
export username="Biswash3"
export password="test123"
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
    -d '{"username":"'"$username"'","password":"'"$password"'"}'
```

```sh
# Get token for registered user
export accountId="d752d8fc-a86d-4b33-85df-0a7615548e14"
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"'"$username"'","password":"'"$password"'"}'
```
{"sessionId":"af6a31cd-f0f9-487b-ae4e-8e177ae39672","accountId":"538b02fa-1586-4b24-ac27-e2cca4f2bfe2","createdAt":"2025-06-03T22:31:28.157129109-04:00","updatedAt":"2025-06-03T22:31:28.157129209-04:00"}

```sh
export token="702bec0e-a87d-41a7-86ef-3270abd6b4c0"
curl http://localhost:8080/api/$accountId/details \
  -H "Authorization: $token"
{"accountId":"538b02fa-1586-4b24-ac27-e2cca4f2bfe2","rewardsBalance":"0"}
```


```sh
curl -X PUT http://localhost:8080/api/update/balance/d752d8fc-a86d-4b33-85df-0a7615548e14 \
  -H "Content-Type: application/json" \
  -H "Authorization: $token" \
  -d '{
    "balance": 100.00,
    "description": "Test transaction",
    "merchantCode": "ABC123"
}'
```