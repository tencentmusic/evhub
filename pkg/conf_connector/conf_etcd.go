package conf_connector

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tencentmusic/evhub/pkg/util/routine"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/tencentmusic/evhub/pkg/log"
)

// EtcdConf is type of configuration
type EtcdConf struct {
	// watch is the name of the watch function
	watch func(c interface{}, eventType int32) error
	// c is the name of the etcd configuration
	c EtcdConfig
	// etcdCli is the name of the etcd client
	etcdCli *clientv3.Client
	// configGroupName is the name config group
	configGroupName string
}

// NewEtcd creates a new Etcd configuration
func NewEtcd(c EtcdConfig, configGroupName string) (*EtcdConf, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   c.EtcdEndpoints,
		DialTimeout: c.EtcdDialTimeout,
	})
	if err != nil {
		return nil, err
	}

	r := &EtcdConf{
		c:               c,
		etcdCli:         cli,
		configGroupName: configGroupName,
	}
	return r, nil
}

// SetWatch is used to watch the configuration
func (s *EtcdConf) SetWatch(watch func(c interface{}, eventType int32) error) error {
	s.watch = watch
	rch := s.etcdCli.Watch(context.Background(), s.configGroupName, clientv3.WithPrefix())
	routine.GoWithRecovery(func() {
		for wresp := range rch {
			for _, ev := range wresp.Events {
				log.Infof("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
				eventType := EventTypeAdd
				if ev.Type == clientv3.EventTypeDelete {
					eventType = EventTypeDelete
				}
				if err := s.watch(ev.Kv.Value, eventType); err != nil {
					log.Errorf("watch err:%v", err)
					continue
				}
			}
		}
	})
	return nil
}

// GetAllConfInfo returns all configuration information
func (s *EtcdConf) GetAllConfInfo() ([]interface{}, error) {
	resp, err := s.etcdCli.Get(context.Background(), s.configGroupName,
		clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
	if err != nil {
		return nil, err
	}
	confInfoList := make([]interface{}, 0, len(resp.Kvs))
	for _, ev := range resp.Kvs {
		log.Infof("confInfo:%v", ev)
		confInfoList = append(confInfoList, ev.Value)
	}
	return confInfoList, nil
}

// SetConfInfo is used to set the configuration information
func (s *EtcdConf) SetConfInfo(key string, confInfo interface{}) error {
	b, err := json.Marshal(confInfo)
	if err != nil {
		return err
	}
	_, err = s.etcdCli.Put(context.Background(), key, string(b))
	if err != nil {
		return err
	}
	return nil
}

// GetConfInfo is used to get the configuration information
func (s *EtcdConf) GetConfInfo(key string) ([]byte, error) {
	resp, err := s.etcdCli.Get(context.Background(), key)
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, fmt.Errorf("not found")
	}

	return resp.Kvs[0].Value, nil
}
