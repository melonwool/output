package output

import (
	"bytes"
	"time"

	"github.com/getsentry/sentry-go"

	"github.com/spf13/viper"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
)

type (
	Watcher struct {
		option Option
	}
	Pair struct {
		Value []byte
	}
	Option struct {
		Addr       string
		Token      string
		DataCenter string
	}
	OptFunc func(option *Option)
	//WatchCallback func(data *Pair)
	WatchCallback func(data []byte)
)

// NewWatcher
func NewWatcher(optFunc ...OptFunc) *Watcher {
	option := Option{}
	for _, fn := range optFunc {
		fn(&option)
	}
	return &Watcher{
		option: option,
	}
}

// WithAddr 添加consul 地址
func WithAddr(addr string) OptFunc {
	return func(option *Option) {
		option.Addr = addr
	}
}

// WIthToken 添加Token验证
func WithToken(token string) OptFunc {
	return func(option *Option) {
		option.Token = token
	}
}

// WithDC 添加datacenter
func WithDC(dc string) OptFunc {
	return func(option *Option) {
		option.DataCenter = dc
	}
}

func ReloadViper(vip *viper.Viper) WatchCallback {
	return func(data []byte) {
		if err := vip.ReadConfig(bytes.NewBuffer(data)); err != nil {
			sentry.CaptureException(err)
		}
		if err := vip.UnmarshalKey("error", &_output); err != nil {
			sentry.CaptureException(err)
		}
	}
}

// WatchKey 监听数据
//func (w *Watcher) WatchKey(key string, pair *Pair) {
func (w *Watcher) WatchKey(key string, callback WatchCallback) {
	var (
		err    error
		params map[string]interface{}
		plan   *watch.Plan
		ch     chan int
	)
	ch = make(chan int, 1)

	params = make(map[string]interface{})
	params["type"] = "key"
	params["key"] = key
	params["datacenter"] = w.option.DataCenter
	params["token"] = w.option.Token
	params["stale"] = true
	plan, err = watch.Parse(params)
	if err != nil {
		// 初始化时
		panic(err)
	}
	go func() {
		// consul 地址
		config := &api.Config{WaitTime: time.Second * 10}
		if err = plan.RunWithConfig(w.option.Addr, config); err != nil {
			panic(err)
		}
	}()
	plan.Handler = func(index uint64, result interface{}) {
		if kvPair, ok := result.(*api.KVPair); ok {
			//pair.Value = kvPair.Value
			//callback(pair)
			callback(kvPair.Value)
			ch <- 1
		}
	}
	go func() {
		for range ch {
		}
	}()
}
