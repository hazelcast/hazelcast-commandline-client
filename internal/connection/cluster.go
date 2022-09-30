package connection

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/config"
	hzcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal/constants"
)

var InvalidStateErr = errors.New("invalid new state")

type RESTCall struct {
	url    string
	params string
}

func CallClusterOperation(config *hazelcast.Config, operation string) (*string, error) {
	var str string
	return CallClusterOperationWithState(config, operation, &str)
}

func CallClusterOperationWithState(config *hazelcast.Config, operation string, state *string) (*string, error) {
	obj, err := NewRESTCall(config, operation, *state)
	if err != nil {
		if errors.Is(err, InvalidStateErr) {
			err = hzcerrors.NewLoggableError(err, "Invalid new state. It should be one the following: %s, %s, %s, %s\n", constants.ClusterStateActive, constants.ClusterStateFrozen, constants.ClusterStateNoMigration, constants.ClusterStatePassive)
		}
		return nil, err
	}
	params := obj.params
	urlStr := obj.url
	pr := strings.NewReader(params)
	var resp *http.Response
	tr := &http.Transport{
		TLSClientConfig: config.Cluster.Network.SSL.TLSConfig(),
	}
	client := &http.Client{Transport: tr}
	switch operation {
	case constants.ClusterGetState, constants.ClusterChangeState, constants.ClusterShutdown:
		resp, err = client.Post(urlStr, "application/x-www-form-urlencoded", pr)
	case constants.ClusterVersion:
		resp, err = client.Get(urlStr)
	}
	if err != nil {
		if msg, handled := hzcerrors.TranslateError(err, config.Cluster.Cloud.Enabled, operation); handled {
			return nil, hzcerrors.NewLoggableError(err, msg)
		}
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, hzcerrors.NewLoggableError(err, "Could not read the response from the cluster")
	}
	sb := string(body)
	return &sb, nil
}

func NewRESTCall(conf *hazelcast.Config, operation string, state string) (*RESTCall, error) {
	var member, url string
	var params string
	//todo iterate over all addresses
	member = config.GetClusterAddress(conf)
	scheme := "http"
	if conf.Cluster.Network.SSL.Enabled == true {
		scheme = "https"
	}
	switch operation {
	case constants.ClusterGetState:
		url = fmt.Sprintf("%s://%s%s", scheme, member, constants.ClusterGetStateEndpoint)
	case constants.ClusterChangeState:
		if !validateState(state) {
			return nil, InvalidStateErr
		}
		url = fmt.Sprintf("%s://%s%s", scheme, member, constants.ClusterChangeStateEndpoint)
	case constants.ClusterShutdown:
		url = fmt.Sprintf("%s://%s%s", scheme, member, constants.ClusterShutdownEndpoint)
	case constants.ClusterVersion:
		url = fmt.Sprintf("%s://%s%s", scheme, member, constants.ClusterVersionEndpoint)
	default:
		panic("Invalid operation to set connection obj.")
	}
	params = newParams(conf, operation, state)
	return &RESTCall{url: url, params: params}, nil
}

func newParams(config *hazelcast.Config, operation string, state string) string {
	var params string
	switch operation {
	case constants.ClusterGetState, constants.ClusterShutdown:
		params = fmt.Sprintf("%s&%s", config.Cluster.Name, config.Cluster.Security.Credentials.Password)
	case constants.ClusterChangeState:
		params = fmt.Sprintf("%s&%s&%s", config.Cluster.Name, config.Cluster.Security.Credentials.Password, state)
	case constants.ClusterVersion:
		params = ""
	default:
		panic("invalid operation to set params.")
	}
	return params
}

func validateState(state string) bool {
	switch strings.ToLower(state) {
	case constants.ClusterStateActive, constants.ClusterStateFrozen, constants.ClusterStateNoMigration, constants.ClusterStatePassive:
		return true
	}
	return false
}
