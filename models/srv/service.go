package srv

import (
	"os"
	"time"

	"github.com/souliot/gateway/master"
	"github.com/souliot/gateway/service"
	logs "github.com/souliot/siot-log"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	etcdTimeout = 10 * time.Second
)

type RegService struct {
	etcdEndpoints []string
	cli           *clientv3.Client
	meta          *master.ServiceMeta
	ser           *service.Service
}

func NewRegService(etcdEndpoints []string, meta *master.ServiceMeta) (rs *RegService, err error) {
	cli, err := master.GetEtcdClient(etcdEndpoints, etcdTimeout)
	if err != nil {
		logs.Error("服务注册，获取Etcd Client 错误：", err)
		os.Exit(0)
		return
	}
	rs = &RegService{
		etcdEndpoints: etcdEndpoints,
		cli:           cli,
		meta:          meta,
	}
	return
}

func (s *RegService) Start() {
	ser, err := service.Register(s.meta, 10, s.cli)
	if err != nil {
		logs.Error("服务注册错误：", err)
		os.Exit(0)
		return
	}
	s.ser = ser
	logs.Info("服务注册成功,注册信息：", s.meta)
	return
}
func (s *RegService) Stop() {
	s.ser.Stop()
	logs.Info("取消服务注册")
	return
}
