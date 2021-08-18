package internal

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/hazelcast/hazelcast-go-client"
)

type RESTCall struct {
	url    string
	params string
}

func CallClusterOperation(config *hazelcast.Config, operation string, state *string) (*string, error) {
	var err error
	obj := NewRESTCall(config, operation, state)
	params := obj.params
	url := obj.url
	pr := strings.NewReader(params)
	var resp *http.Response
	switch operation {
	case ClusterGetState, ClusterChangeState, ClusterShutdown:
		resp, err = http.Post(url, "application/x-www-form-urlencoded", pr)
	case ClusterQuery:
		resp, err = http.Get(url)
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	sb := string(body)
	return &sb, nil
}

func NewRESTCall(config *hazelcast.Config, operation string, state *string) *RESTCall {
	var member, url string
	var params string
	var addresses []string = config.Cluster.Network.Addresses
	rand.Seed(time.Now().Unix())
	member = addresses[rand.Intn(len(addresses))]
	switch operation {
	case ClusterGetState:
		url = fmt.Sprintf("http://%s%s", member, ClusterGetStateEndpoint)
	case ClusterChangeState:
		url = fmt.Sprintf("http://%s%s", member, ClusterChangeStateEndpoint)
	case ClusterShutdown:
		url = fmt.Sprintf("http://%s%s", member, ClusterShutdownEndpoint)
	case ClusterQuery:
		url = fmt.Sprintf("http://%s%s", member, ClusterQueryEndpoint)
	default:
		panic("Invalid operation to set connection obj.")
	}
	params = NewParams(config, operation, state)
	return &RESTCall{url: url, params: params}
}

func NewParams(config *hazelcast.Config, operation string, state *string) string {
	var params string
	switch operation {
	case ClusterGetState, ClusterShutdown:
		params = fmt.Sprintf("%s&%s", config.Cluster.Name, config.Cluster.Security.Credentials.Password)
	case ClusterChangeState:
		params = fmt.Sprintf("%s&%s&%s", config.Cluster.Name, config.Cluster.Security.Credentials.Password, EnsureState(state))
	case ClusterQuery:
		params = ""
	default:
		panic("invalid operation to set params.")
	}
	return params
}

func EnsureState(state *string) string {
	switch *state {
	case ClusterStateActive, ClusterStateFrozen, ClusterStateNoMigration, ClusterStatePassive:
		return *state
	default:
		panic("invalid new state.")
	}
}
