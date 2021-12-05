package master

import (
	"context"
	"sync"

	"go.etcd.io/etcd/client/v3/concurrency"

	"github.com/lanceryou/Registry/pkg/log"
	"github.com/lanceryou/Registry/pkg/waitgroup"
)

func New(opts ...MasterOption) (*RegistryMaster, error) {
	var opt Options
	for _, o := range opts {
		o(&opt)
	}

	rm := &RegistryMaster{
		quit: make(chan struct{}),
	}

	rm.Wrap(rm.masterLoop)
	return rm, nil
}

/**
 * master 的主要作用
 * 负责全局调度，负载均衡
 * 功能：master抢占election，接受请求返回master地址
 **/
type RegistryMaster struct {
	sync.RWMutex
	waitgroup.WaitGroupWrapper
	masterEndPoint string
	addr           string
	quit           chan struct{}
	stop           chan struct{}
	election       *concurrency.Election
}

func (r *RegistryMaster) Stop() {
	close(r.quit)
	r.Wait()
}

func (r *RegistryMaster) masterLoop() {
	for {
		select {
		case <-r.quit:
			return
		default:
		}
		err := r.election.Campaign(context.TODO(), r.addr)
		if err != nil {
			log.Errorf("election error %v", err)
			continue
		}

		// 选主成功
		log.Infof("election success %v", r.addr)
		r.Lock()
		r.masterEndPoint = r.addr
		r.Unlock()

		select {
		case kv := <-r.election.Observe(context.TODO()):
			r.Lock()
			r.masterEndPoint = string(kv.Kvs[0].Value)
			r.Unlock()
		}
	}
}

func (r *RegistryMaster) isMaster() bool {
	var me string
	r.RLock()
	me = r.masterEndPoint
	r.RUnlock()

	return me == r.masterEndPoint
}
