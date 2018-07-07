package loadbalance

import (
	"strings"

	"github.com/k0kubun/pp"
	"github.com/rai-project/config"
	"github.com/rai-project/vipertags"
)

type loadbalanceConfig struct {
	Enabled bool          `json:"enabled" config:"loadbalance.enabled" `
	Method  string        `json:"method" config:"loadbalance.method"`
	done    chan struct{} `json:"-" config:"-"`
}

var (
	Config = &loadbalanceConfig{
		done: make(chan struct{}),
	}
)

func (*loadbalanceConfig) ConfigName() string {
	return "S3"
}

func (a *loadbalanceConfig) SetDefaults() {
	vipertags.SetDefaults(a)
}

func (a *loadbalanceConfig) Read() {
	defer close(a.done)
	vipertags.Fill(a)
	a.Method = strings.ToLower(a.Method)
	if a.Method == "" || a.Method == "default" {
		a.Method = "roundrobin"
	}
	switch a.Method {
	case "bounded", "consistent", "p2c", "r2", "roundrobin":
		break
	default:
		panic("invalid load balancing method " + a.Method)
	}
}

func (c loadbalanceConfig) Wait() {
	<-c.done
}

func (c *loadbalanceConfig) String() string {
	return pp.Sprintln(c)
}

func (c *loadbalanceConfig) Debug() {
	log.Debug("S3 Config = ", c)
}

func init() {
	config.Register(Config)
}
