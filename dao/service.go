package dao

import (
	"errors"
	"gateway/dto"
	"gateway/golang_common/lib"
	"gateway/public"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"strings"
	"sync"
)

type ServiceDetail struct {
	Info          *ServiceInfo   `json:"info" description:"基本信息"`
	HTTPRule      *HttpRule      `json:"http_rule" description:"http_rule"`
	TCPRule       *TcpRule       `json:"tcp_rule" description:"tcp_rule"`
	GRPCRule      *GrpcRule      `json:"grpc_rule" description:"grpc_rule"`
	LoadBalance   *LoadBalance   `json:"load_balance" description:"load_balance"`
	AccessControl *AccessControl `json:"access_control" description:"access_control"`
}

var ServiceManagerHandler *ServiceManager

func init() {
	ServiceManagerHandler = NewServiceManager()
}

type ServiceManager struct {
	ServiceMap   map[string]*ServiceDetail
	ServiceSlice []*ServiceDetail
	Locker       sync.RWMutex
	init         sync.Once
	err          error
}

func NewServiceManager() *ServiceManager {
	return &ServiceManager{
		ServiceMap:   make(map[string]*ServiceDetail),
		ServiceSlice: make([]*ServiceDetail, 0),
		Locker:       sync.RWMutex{},
		init:         sync.Once{},
	}
}

func (s *ServiceManager) HTTPAccessMode(c *gin.Context) (*ServiceDetail, error) {
	//1、前缀匹配 /abc ==> serviceSlice.rule
	//2、域名匹配 www.test.com ==> serviceSlice.rule
	//host c.Request.Host
	//path c.Request.URL.Path
	host := c.Request.Host
	host = host[0:strings.Index(host, ":")]
	path := c.Request.URL.Path
	for _, serviceItem := range s.ServiceSlice {
		if serviceItem.Info.LoadType != public.LoadTypeHTTP {
			continue
		}
		if serviceItem.HTTPRule.RuleType == public.HTTPRuleTypeDomain {
			if serviceItem.HTTPRule.Rule == host {
				return serviceItem, nil
			}
		}
		if serviceItem.HTTPRule.RuleType == public.HTTPRuleTypePrefixURL {
			if strings.HasPrefix(path, serviceItem.HTTPRule.Rule) {
				return serviceItem, nil
			}
		}
	}
	return nil, errors.New("not matched service")
}

func (sm *ServiceManager) LoadOnce() error {
	sm.init.Do(func() {
		serviceInfo := &ServiceInfo{}
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		tx, err := lib.GetGormPool("default")
		if err != nil {
			sm.err = err
			return
		}
		params := &dto.ServiceListInput{PageNum: 1, PageSize: 99999}
		list, _, err := serviceInfo.PageList(c, tx, params)
		if err != nil {
			sm.err = err
			return
		}
		sm.Locker.Lock()
		defer sm.Locker.Unlock()
		for _, listItem := range list {
			tmpItem := listItem
			serviceDetail, err := tmpItem.ServiceDetail(c, tx, &tmpItem)
			//fmt.Println("serviceDetail")
			//fmt.Println(public.Obj2Json(serviceDetail))
			if err != nil {
				sm.err = err
				return
			}
			sm.ServiceMap[listItem.ServiceName] = serviceDetail
			sm.ServiceSlice = append(sm.ServiceSlice, serviceDetail)
		}
	})
	return sm.err
}
