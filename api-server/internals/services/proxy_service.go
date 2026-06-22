package services

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/uddinArsalan/devdeploy/internals/utils"
)

type ProxyService struct {
	portMap *utils.PortMap
}

func NewProxyService(portMap *utils.PortMap) *ProxyService {
	return &ProxyService{
		portMap: portMap,
	}
}

func (ps *ProxyService) Route(hostname string) (*url.URL, error) {
	domain, ok := ps.portMap.PortMapping[hostname]
	fmt.Printf("\nPORT MAP DOMAIN\n%v",domain)
	if !ok {
		return nil, errors.New("404 not found")
	}

	return url.Parse(
		fmt.Sprintf("http://localhost:%v", domain.Port),
	)
}
