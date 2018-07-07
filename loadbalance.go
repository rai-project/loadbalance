package loadbalance

import (
	"sync"
	"time"

	"github.com/lafikl/liblb/bounded"
	"github.com/lafikl/liblb/consistent"
	"github.com/lafikl/liblb/p2c"
	"github.com/lafikl/liblb/r2"
	"github.com/pkg/errors"
	"golang.org/x/sync/syncmap"
)

type LoadBalancer interface {
	Get(name string) (string, error)
	Done(resource string) error
}

type balancer struct {
	sync.Mutex
	resources []string
	used      syncmap.Map
	history   syncmap.Map
	alg       LoadBalancer
}

type balancerInfo struct {
	Resources []string               `json:"resources,omitempty"`
	Used      map[string]bool        `json:"used,omitempty"`
	History   map[string][]time.Time `json:"history,omitempty"`
	Algorithm string                 `json:"algorithm,omitempty"`
}

type boundedToKeyBalancer struct {
	*bounded.Bounded
}

func (b *boundedToKeyBalancer) Get(name string) (string, error) {
	return b.Balance(name)
}
func (b *boundedToKeyBalancer) Done(string) error {
	return nil
}

type consistentToKeyBalancer struct {
	*consistent.Consistent
}

func (b *consistentToKeyBalancer) Get(name string) (string, error) {
	return b.Balance(name)
}
func (b *consistentToKeyBalancer) Done(string) error {
	return nil
}

type p2cToKeyBalancer struct {
	*p2c.P2C
}

func (b *p2cToKeyBalancer) Get(name string) (string, error) {
	return b.Balance(name)
}
func (b *p2cToKeyBalancer) Done(string) error {
	return nil
}

type r2ToKeyBalancer struct {
	*r2.R2
}

func (b *r2ToKeyBalancer) Get(name string) (string, error) {
	return b.Balance()
}
func (b *r2ToKeyBalancer) Done(string) error {
	return nil
}

func New(args ...string) (LoadBalancer, error) {
	if !Config.Enabled {
		return nil, errors.New("the load balancer is disabled")
	}

	var alg LoadBalancer
	var err error
	switch Config.Method {
	case "bounded":
		alg = &boundedToKeyBalancer{bounded.New(args...)}
	case "consistent":
		alg = &consistentToKeyBalancer{consistent.New(args...)}
	case "p2c":
		alg = &p2cToKeyBalancer{p2c.New(args...)}
	case "r2", "roundrobin":
		alg = &r2ToKeyBalancer{r2.New(args...)}
	default:
		err = errors.Errorf("invalid load balancing method %s", Config.Method)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create load balancer %s with %v", Config.Method, args)
	}
	var used syncmap.Map
	var history syncmap.Map
	for arg := range args {
		used.Store(arg, false)
		history.Store(arg, []time.Time{})
	}
	return &balancer{alg: alg, resources: args, used: used, history: history}, nil
}

func (b *balancer) AllUsed() bool {
	res := true
	b.used.Range(func(key0, val0 interface{}) bool {
		val := val0.(bool)
		if val == false {
			res = false
			return false
		}
		return true
	})
	return res
}

func (b *balancer) Get(name string) (string, error) {
	b.Lock()
	defer b.Unlock()

	if b.AllUsed() {
		return "", errors.New("all resources are currently in use")
	}

	resource, err := b.alg.Get(name)
	if err != nil {
		return "", errors.Wrapf(err, "unable to balance")
	}

	b.used.Store(resource, true)
	history, _ := b.history.Load(resource)
	b.history.Store(resource, append(history.([]time.Time), time.Now()))

	return resource, nil
}

func (b *balancer) Done(resource string) error {
	b.Lock()
	defer b.Unlock()

	if resource == "" {
		return errors.New("invalid empty resource name")
	}

	b.used.Store(resource, false)

	return nil
}

func (b *balancer) Info() (balancerInfo, error) {
	b.Lock()
	defer b.Unlock()

	bl := *b
	var info balancerInfo
	info.Algorithm = Config.Method
	info.Resources = bl.resources
	bl.used.Range(func(key0, val0 interface{}) bool {
		key := key0.(string)
		val := val0.(bool)
		info.Used[key] = val
		return true
	})
	bl.history.Range(func(key0, val0 interface{}) bool {
		key := key0.(string)
		val := val0.([]time.Time)
		info.History[key] = val
		return true
	})
	return info, nil
}
