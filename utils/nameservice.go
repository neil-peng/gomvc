package utils

import (
	"errors"
	"strings"

	"github.com/neil-peng/gomvc/conf"
)

type NameService interface {
	GetServer(service string) (ip string, port string, err error)
}

type IpNameService struct{}

var IpServer IpNameService

//example: (&IpNameService{}).GetServer("127.0.0.1:3600")
func (i *IpNameService) GetServer(ipAndPort string) (string, string, error) {
	addrs := strings.Split(ipAndPort, ":")
	if len(addrs) == 2 {
		return addrs[0], addrs[1], nil
	}
	return "", "", errors.New(conf.ERROR_NAMESERVICE_ERROR)
}
