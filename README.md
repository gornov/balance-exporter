### Usage of balance_exporter:

```
balance_exporter
  -address string
        Address on which to expose metrics. (default ":9913")
  -metrics_endpoint string
        Path under which to expose metrics. (default "/metrics")
  -metrics_namespace string
        Prometheus metrics namespace (default "wallet")
  -scrape_timeout int
        The number of seconds to wait for an HTTP response from the scrape_timeout (default 60)
  -scrape_uri string
        URI to api page (default "http://localhost/api/WalletsClientBalances/0")
```

### Example of run container command

```
docker run --rm nf404/balance-exporter -scrape_uri http://localhost/api/WalletsClientBalances/0000
```
