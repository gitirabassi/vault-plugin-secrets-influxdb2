# vault-plugin-secrets-influxdb2



## Development

```shell

docker run -p 9999:9999 --rm influxdb/influxdb:2.0.0-beta

make dev &

export VAULT_ADDR=http://localhost:8200
export VAULT_TOKEN=root

vault status

vault write influxdb2/config host=http://localhost:9999 initialize=true
vault read influxdb2/config
vault write influxdb2/roles/first org_id=12114231kjh permissions=asdasdaks,asdadsasda,asdasdasd
vault read influxdb2/roles/first
vault read influxdb2/creds/first
```

