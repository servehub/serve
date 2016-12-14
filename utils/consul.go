package utils

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/fatih/color"
	consul "github.com/hashicorp/consul/api"
)

func ConsulClient(consulAddress string) (*consul.Client, error) {
	conf := consul.DefaultConfig()
	conf.Address = consulAddress
	return consul.NewClient(conf)
}

var PutConsulKv = func(client *consul.Client, key string, value string) error {
	log.Printf("consul put `%s`: %s", key, value)
	_, err := client.KV().Put(&consul.KVPair{Key: strings.TrimPrefix(key, "/"), Value: []byte(value)}, nil)
	return err
}

func ListConsulKv(client *consul.Client, prefix string, q *consul.QueryOptions) (consul.KVPairs, error) {
	log.Printf("consul list `%s`", prefix)
	list, _, err := client.KV().List(prefix, q)
	return list, err
}

var DelConsulKv = func(client *consul.Client, key string) error {
	log.Printf("consul delete `%s`", key)
	_, err := client.KV().Delete(strings.TrimPrefix(key, "/"), nil)
	return err
}

var RegisterPluginData = func(plugin string, packageName string, data string, consulAddress string) error {
	consulApi, err := ConsulClient(consulAddress)
	if err != nil {
		return err
	}

	return PutConsulKv(consulApi, "services/data/"+packageName+"/"+plugin, data)
}

var DeletePluginData = func(plugin string, packageName string, consulAddress string) error {
	log.Println(color.YellowString("Delete %s for %s package in consul", plugin, packageName))
	consulApi, err := ConsulClient(consulAddress)
	if err != nil {
		return err
	}

	return DelConsulKv(consulApi, "services/data/"+packageName+"/"+plugin)
}

func MarkAsOutdated(client *consul.Client, name string, delay time.Duration) error {
	log.Printf("Mark service `%s` as outdated\n", name)
	json := fmt.Sprintf(`{"endOfLife":%d}`, time.Now().Add(delay).UnixNano()/int64(time.Millisecond))
	return PutConsulKv(client, "services/outdated/"+name, json)
}
