package internal

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/cobra"
)

type ConnectionObj struct {
	url    *string
	params *string
}

func ClusterConnect(cmd *cobra.Command, operation string, state *string) string {
	config, err := RetrieveFlagValues(cmd)
	if err != nil {
		log.Fatal(err)
	}
	obj := SetConnectionObj(config, operation, state)
	params := obj.params
	url := obj.url
	pr := strings.NewReader(*params)
	var resp *http.Response
	switch operation {
	case ClusterGetState, ClusterChangeState, ClusterShutdown:
		resp, err = http.Post(*url, "application/x-www-form-urlencoded", pr)
	case ClusterQuery:
		resp, err = http.Get(*url)
	}
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	sb := string(body)
	return sb
}

func SetConnectionObj(config *hazelcast.Config, operation string, state *string) *ConnectionObj {
	var member, _url string
	var _params *string
	var addresses []string = config.Cluster.Network.Addresses
	rand.Seed(time.Now().Unix())
	member = addresses[rand.Intn(len(addresses))]
	switch operation {
	case ClusterGetState:
		_url = fmt.Sprintf("http://%s%s", member, ClusterGetStateEndpoint)
	case ClusterChangeState:
		_url = fmt.Sprintf("http://%s%s", member, ClusterChangeStateEndpoint)
	case ClusterShutdown:
		_url = fmt.Sprintf("http://%s%s", member, ClusterShutdownEndpoint)
	case ClusterQuery:
		_url = fmt.Sprintf("http://%s%s", member, ClusterQueryEndpoint)
	default:
		panic("Invalid operation to set connection obj.")
	}
	_params = SetParams(config, operation, state)
	return &ConnectionObj{url: &_url, params: _params}
}

func SetParams(config *hazelcast.Config, operation string, state *string) *string {
	var params string
	switch operation {
	case ClusterGetState, ClusterShutdown:
		params = fmt.Sprintf("%s&%s", config.Cluster.Name, config.Cluster.Security.Credentials.Password)
	case ClusterChangeState:
		params = fmt.Sprintf("%s&%s&%s", config.Cluster.Name, config.Cluster.Security.Credentials.Password, SetState(state))
	case ClusterQuery:
		params = ""
	default:
		panic("Invalid operation to set params.")
	}
	return &params
}

func SetState(state *string) string {
	switch *state {
	case ClusterStateActive, ClusterStateFrozen, ClusterStateNoMigration, ClusterStatePassive:
		return *state
	default:
		panic("Invalid new state.")
	}
}
