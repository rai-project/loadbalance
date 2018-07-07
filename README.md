# Load Balancer [![Build Status](https://travis-ci.org/rai-project/loadbalance.svg?branch=master)](https://travis-ci.org/rai-project/loadbalance)

## Config

```yaml
loadbalance:
  enabled: true | false (default to false)
  method: "bounded" | "consistent" | "p2c" | "r2" | "roundrobin" (default to "roundrobin")
```

An example config is

```yaml
loadbalance:
  enabled: true
  method: "roundrobin"
```
