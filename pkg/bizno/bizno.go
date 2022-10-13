package bizno

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/sony/sonyflake"
	"github.com/tencentmusic/evhub/pkg/log"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

// Generator is the interface for creating a new biz number generator
type Generator interface {
	NextID() (string, error)
}

// bizNoGenerator is the biz number generator
type bizNoGenerator struct {
	// nodeNum is the count of generator nodes
	nodeNum int64
	// flake is the name of the snowflake generator
	flake *sonyflake.Sonyflake
	// localIP is the name of the local ip
	localIP string
	// etcdCli is the name of the etcd client
	etcdCli *clientv3.Client
}

var (
	KeyGeneratorNodeNum = "/evhub/generator/node/num"
	KeyGeneratorNode    = "/evhub/generator/node/num/%v"
)

const (
	// MaxWorkID is the maximum number of workers
	MaxWorkID = 65536
)

// KeyGeneratorNodeKey returns the key generator node for the given	node number
func KeyGeneratorNodeKey(nodeNum int64) string {
	return fmt.Sprintf(KeyGeneratorNode, nodeNum)
}

// NewGenerator creates a new generator for the given configuration
func NewGenerator(etcdCli *clientv3.Client) (Generator, error) {
	s := &bizNoGenerator{
		etcdCli: etcdCli,
	}

	nodeNum, err := s.getNodeNum()
	if err != nil {
		return nil, err
	}
	s.nodeNum = nodeNum
	log.Infof("generator node num:%d ip:%v", s.nodeNum, s.localIP)

	flake := sonyflake.NewSonyflake(sonyflake.Settings{
		MachineID: func() (uint16, error) {
			return uint16(s.nodeNum % MaxWorkID), nil
		},
	})

	s.flake = flake
	return s, nil
}

// NextID is used to get the next biz number
func (s *bizNoGenerator) NextID() (string, error) {
	idUint, err := s.flake.NextID()
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(int64(idUint), 10), nil
}

// getNodeNum is used to get the node number
func (s *bizNoGenerator) getNodeNum() (int64, error) {
	var nodeNum int64
	_, err := concurrency.NewSTM(s.etcdCli, func(stm concurrency.STM) error {
		num := stm.Get(KeyGeneratorNodeNum)
		if num != "" {
			var err error
			nodeNum, err = strconv.ParseInt(num, 10, 60)
			if err != nil {
				return err
			}
		}
		nodeNum++
		stm.Put(KeyGeneratorNodeNum, strconv.FormatInt(nodeNum, 10))
		ip, err := getLocalIP()
		if err != nil {
			return err
		}
		s.localIP = ip

		seession, err := concurrency.NewSession(s.etcdCli)
		if err != nil {
			return err
		}
		if err := concurrency.NewMutex(seession, KeyGeneratorNodeKey(nodeNum)).Lock(context.Background()); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return 0, err
	}
	return nodeNum, nil
}

// getLocalIP returns the local IP address
func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", errors.New("not found ip")
}
