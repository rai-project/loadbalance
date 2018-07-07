package loadbalance

import (
	"github.com/lafikl/liblb/bounded"
	"github.com/lafikl/liblb/consistent"
	"github.com/lafikl/liblb/p2c"
	"github.com/lafikl/liblb/r2"
	"github.com/pkg/errors"
)

type LoadBalancer interface {
	Balance(key string) (string, error)
}

type balancer struct {
	alg LoadBalancer
}

type r2ToKeyBalancer struct {
	*r2.R2
}

func (b *r2ToKeyBalancer) Balance(key string) (string, error) {
	return b.R2.Balance()
}

func New(args ...string) (LoadBalancer, error) {
	if !Config.Enabled {
		return nil, errors.New("the load balancer is disabled")
	}

	var alg LoadBalancer
	var err error
	switch Config.Method {
	case "bounded":
		alg = bounded.New(args...)
	case "consistent":
		alg = consistent.New(args...)
	case "p2c":
		alg = p2c.New(args...)
	case "r2", "roundrobin":
		alg = &r2ToKeyBalancer{
			r2.New(args...),
		}
	default:
		err = errors.Errorf("invalid load balancing method %s", Config.Method)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create load balancer %s with %v", Config.Method, args)
	}
	return &balancer{alg: alg}, nil
}

func (b *balancer) Balance(arg string) (string, error) {
	return b.alg.Balance(arg)
}
