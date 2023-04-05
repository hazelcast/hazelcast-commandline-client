/*
 * Copyright (c) 2008-2023, Hazelcast, Inc. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License")
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package it

import (
	"context"
	"crypto/tls"
	"fmt"
	"math/rand"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
	hz "github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/logger"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

const (
	EnvEnableTraceLogging = "ENABLE_TRACE"
	EnvMemberCount        = "MEMBER_COUNT"
	EnvEnableSSL          = "ENABLE_SSL"
	EnvHzVersion          = "HZ_VERSION"
)

func DefaultClusterName() string {
	return NewUniqueObjectName("clc-test")
}

var defaultClusterName = DefaultClusterName()
var rc *RemoteControllerClientWrapper
var rcMu = &sync.RWMutex{}
var defaultTestCluster = NewSingletonTestCluster(defaultClusterName, func() *TestCluster {
	port := NextPort()
	return rc.startNewCluster(MemberCount(), xmlConfig(defaultClusterName, port), port)
})
var idGen = ReferenceIDGenerator{}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func NewUniqueObjectName(service string, labels ...string) string {
	ls := strings.Join(labels, "_")
	if ls != "" {
		ls = fmt.Sprintf("-%s", ls)
	}
	// make sure the random part is at least 4 characters long
	return fmt.Sprintf("test-%s-%d-%d%s", service, idGen.NextID(), rand.Intn(100_000)+1000, ls)
}

func TraceLoggingEnabled() bool {
	return os.Getenv(EnvEnableTraceLogging) == "1"
}

func SSLEnabled() bool {
	return os.Getenv(EnvEnableSSL) == "1"
}

func HzVersion() string {
	version := os.Getenv(EnvHzVersion)
	if version == "" {
		version = "5.1"
	}
	return version
}

func MemberCount() int {
	if memberCountStr := os.Getenv(EnvMemberCount); memberCountStr != "" {
		memberCount, err := strconv.Atoi(memberCountStr)
		if err != nil {
			panic(err)
		}
		return memberCount
	}
	return 1
}

func CreateDefaultRemoteController() *RemoteControllerClientWrapper {
	return newRemoteControllerClientWrapper(CreateRemoteController("localhost:9701"))
}

func CreateRemoteController(addr string) *RemoteControllerClient {
	transport := check.MustValue(thrift.NewTSocketConf(addr, nil))
	bufferedTransport := thrift.NewTBufferedTransport(transport, 4096)
	protocol := thrift.NewTBinaryProtocolConf(bufferedTransport, nil)
	client := thrift.NewTStandardClient(protocol, protocol)
	rc := NewRemoteControllerClient(client)
	check.Must(transport.Open())
	return rc
}

func ensureRemoteController(launchDefaultCluster bool) *RemoteControllerClientWrapper {
	rcMu.Lock()
	defer rcMu.Unlock()
	if rc == nil {
		rc = CreateDefaultRemoteController()
		if ping, err := rc.Ping(context.Background()); err != nil {
			panic(err)
		} else if !ping {
			panic("remote controller not accesible")
		}
	}
	return rc
}

func StartNewClusterWithOptions(clusterName string, port, memberCount int) *TestCluster {
	ensureRemoteController(false)
	config := xmlConfig(clusterName, port)
	return rc.startNewCluster(memberCount, config, port)
}

func StartNewClusterWithConfig(memberCount int, config string, port int) *TestCluster {
	ensureRemoteController(false)
	return rc.startNewCluster(memberCount, config, port)
}

type RemoteControllerClientWrapper struct {
	mu *sync.Mutex
	rc *RemoteControllerClient
}

func newRemoteControllerClientWrapper(rc *RemoteControllerClient) *RemoteControllerClientWrapper {
	return &RemoteControllerClientWrapper{
		mu: &sync.Mutex{},
		rc: rc,
	}
}

func (rcw *RemoteControllerClientWrapper) startNewCluster(memberCount int, config string, port int) *TestCluster {
	cluster := check.MustValue(rcw.CreateClusterKeepClusterName(context.Background(), HzVersion(), config))
	memberUUIDs := make([]string, 0, memberCount)
	for i := 0; i < memberCount; i++ {
		member := check.MustValue(rcw.StartMember(context.Background(), cluster.ID))
		memberUUIDs = append(memberUUIDs, member.UUID)
	}
	return &TestCluster{
		RC:          rcw,
		ClusterID:   cluster.ID,
		MemberUUIDs: memberUUIDs,
		Port:        port,
	}
}

func (rcw *RemoteControllerClientWrapper) StartMember(ctx context.Context, clusterID string) (*Member, error) {
	rcw.mu.Lock()
	defer rcw.mu.Unlock()
	return rcw.rc.StartMember(ctx, clusterID)
}

func (rcw *RemoteControllerClientWrapper) Ping(ctx context.Context) (bool, error) {
	rcw.mu.Lock()
	defer rcw.mu.Unlock()
	return rcw.rc.Ping(ctx)
}

func (rcw *RemoteControllerClientWrapper) CreateClusterKeepClusterName(ctx context.Context, hzVersion string, xmlconfig string) (*Cluster, error) {
	rcw.mu.Lock()
	defer rcw.mu.Unlock()
	return rcw.rc.CreateClusterKeepClusterName(ctx, hzVersion, xmlconfig)
}

func (rcw *RemoteControllerClientWrapper) ShutdownMember(ctx context.Context, clusterID string, memberID string) (bool, error) {
	rcw.mu.Lock()
	defer rcw.mu.Unlock()
	return rcw.rc.ShutdownMember(ctx, clusterID, memberID)
}

func (rcw *RemoteControllerClientWrapper) TerminateMember(ctx context.Context, clusterID string, memberID string) (bool, error) {
	rcw.mu.Lock()
	defer rcw.mu.Unlock()
	return rcw.rc.TerminateMember(ctx, clusterID, memberID)
}

func (rcw *RemoteControllerClientWrapper) ExecuteOnController(ctx context.Context, clusterID string, script string, lang Lang) (*Response, error) {
	rcw.mu.Lock()
	defer rcw.mu.Unlock()
	return rcw.rc.ExecuteOnController(ctx, clusterID, script, lang)
}

type TestCluster struct {
	RC          *RemoteControllerClientWrapper
	ClusterID   string
	MemberUUIDs []string
	Port        int
}

func (c TestCluster) Shutdown() {
	// TODO: add Terminate method.
	for _, memberUUID := range c.MemberUUIDs {
		c.RC.ShutdownMember(context.Background(), c.ClusterID, memberUUID)
	}
}

func (c TestCluster) DefaultConfig() hz.Config {
	config := hz.Config{}
	config.Cluster.Name = c.ClusterID
	config.Cluster.Network.SetAddresses(fmt.Sprintf("localhost:%d", c.Port))
	if SSLEnabled() {
		config.Cluster.Network.SSL.Enabled = true
		config.Cluster.Network.SSL.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
	}
	if TraceLoggingEnabled() {
		config.Logger.Level = logger.TraceLevel
	}
	return config
}

func (c TestCluster) DefaultConfigWithNoSSL() hz.Config {
	config := hz.Config{}
	config.Cluster.Name = c.ClusterID
	config.Cluster.Network.SetAddresses(fmt.Sprintf("localhost:%d", c.Port))
	if TraceLoggingEnabled() {
		config.Logger.Level = logger.TraceLevel
	}
	return config
}

func (c TestCluster) StartMember(ctx context.Context) (*Member, error) {
	return c.RC.StartMember(ctx, c.ClusterID)
}

func xmlConfig(clusterName string, port int) string {
	return fmt.Sprintf(`
        <hazelcast xmlns="http://www.hazelcast.com/schema/config"
            xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
            xsi:schemaLocation="http://www.hazelcast.com/schema/config
            http://www.hazelcast.com/schema/config/hazelcast-config-4.0.xsd">
            <cluster-name>%s</cluster-name>
            <network>
               <port>%d</port>
            </network>
			<jet enabled="true" />
        </hazelcast>
	`, clusterName, port)
}

func getLoggerLevel() logger.Level {
	if TraceLoggingEnabled() {
		return logger.TraceLevel
	}
	return logger.InfoLevel
}

func getDefaultClient(config *hz.Config) *hz.Client {
	lv := getLoggerLevel()
	if lv == logger.TraceLevel {
		config.Logger.Level = lv
	}
	client, err := hz.StartNewClientWithConfig(context.Background(), *config)
	if err != nil {
		panic(err)
	}
	return client
}

// Eventually asserts that given condition will be met in 2 minutes,
// checking target function every 200 milliseconds.
func Eventually(t *testing.T, condition func() bool, msgAndArgs ...interface{}) {
	if !assert.Eventually(t, condition, time.Minute*2, time.Millisecond*200, msgAndArgs...) {
		t.FailNow()
	}
}

// Never asserts that the given condition doesn't satisfy in 3 seconds,
// checking target function every 200 milliseconds.
func Never(t *testing.T, condition func() bool, msgAndArgs ...interface{}) {
	if !assert.Never(t, condition, time.Second*3, time.Millisecond*200, msgAndArgs) {
		t.FailNow()
	}
}

// WaitEventually waits for the waitgroup for 2 minutes
// Fails the test if 2 mimutes is reached.
func WaitEventually(t *testing.T, wg *sync.WaitGroup) {
	WaitEventuallyWithTimeout(t, wg, time.Minute*2)
}

// WaitEventuallyWithTimeout waits for the waitgroup for the specified max timeout.
// Fails the test if given timeout is reached.
func WaitEventuallyWithTimeout(t *testing.T, wg *sync.WaitGroup, timeout time.Duration) {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-c:
		//done successfully
	case <-timer.C:
		t.FailNow()
	}
}

func EqualStringContent(b1, b2 []byte) bool {
	s1 := sortedString(b1)
	s2 := sortedString(b2)
	return s1 == s2
}

func sortedString(b []byte) string {
	bc := make([]byte, len(b))
	copy(bc, b)
	sort.Slice(bc, func(i, j int) bool {
		return bc[i] < bc[j]
	})
	s := strings.ReplaceAll(string(bc), " ", "")
	s = strings.ReplaceAll(s, "\t", "")
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	return s
}

var nextPort int32 = 10000

func NextPort() int {
	// let minimum step be 10, just a round, safe value.
	// note that some tests may not use the MemberCount() function for the cluster size.
	const maxStep = 10
	step := MemberCount()
	if step < maxStep {
		step = maxStep
	}
nextblock:
	for {
		start := int(atomic.AddInt32(&nextPort, int32(step))) - step
		// check that all ports in the range are open
		for port := start; port < start+step; port++ {
			if !isPortOpen(port) {
				// ignoring the error from fmt.Fprintf, not useful in this case.
				_, _ = fmt.Fprintf(os.Stderr, "it.NextPort: %d is not open, skipping the block: [%d:%d]\n", port, start, start+step)
				continue nextblock
			}
		}
		return start
	}
}

func isPortOpen(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), 10*time.Millisecond)
	if err != nil {
		return true
	}
	// ignoring the error from conn.Close, since there's nothing useful to do with it.
	_ = conn.Close()
	return false
}

type SingletonTestCluster struct {
	mu       *sync.Mutex
	cls      *TestCluster
	launcher func() *TestCluster
	name     string
}

func NewSingletonTestCluster(name string, launcher func() *TestCluster) *SingletonTestCluster {
	return &SingletonTestCluster{
		name:     name,
		mu:       &sync.Mutex{},
		launcher: launcher,
	}
}

type testLogger interface {
	Logf(format string, args ...interface{})
}

func (c *SingletonTestCluster) Launch(t testLogger) *TestCluster {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cls != nil {
		return c.cls
	}
	t.Logf("Launching the auto-shutdown test cluster: %s", c.name)
	c.cls = c.launcher()
	return c.cls
}

type CLCHome struct {
	home string
}

func NewCLCHome() (*CLCHome, error) {
	home, err := os.MkdirTemp("", "clc-")
	if err != nil {
		return nil, err
	}
	return &CLCHome{home: home}, err
}

func (h *CLCHome) Path() string {
	return h.home
}

func (h *CLCHome) Destroy() error {
	if h.home != "" {
		return os.RemoveAll(h.home)
	}
	return nil
}

func (h *CLCHome) WithFile(path string, data []byte, fn func(path string)) {
	path = paths.Join(h.home, path)
	if err := os.WriteFile(path, data, 0600); err != nil {
		panic(fmt.Errorf("writing file: %w", err))
	}
	fn(path)
}

func WithTempFile(fn func(*os.File)) {
	f, err := os.CreateTemp("", "clc-*")
	if err != nil {
		panic(fmt.Errorf("creating temp file: %w", err))
	}
	defer func() {
		// errors are ignored
		f.Close()
		os.Remove(f.Name())
	}()
	fn(f)
}

func WithTempConfigFile(m map[string]any, fn func(path string)) {
	b, err := yaml.Marshal(m)
	if err != nil {
		panic(fmt.Errorf("marhaling YAML: %w", err))
	}
	WithTempFile(func(f *os.File) {
		if err := os.WriteFile(f.Name(), b, 0600); err != nil {
			panic(fmt.Errorf("writing temp file: %w", err))
		}
		fn(f.Name())
	})
}

func ConfigToMap(c hz.Config) map[string]any {
	mc := map[string]any{}
	var addr string
	if len(c.Cluster.Network.Addresses) > 0 {
		addr = c.Cluster.Network.Addresses[0]
	}
	mc["cluster"] = map[string]any{
		"name":    c.Cluster.Name,
		"address": addr,
	}
	if c.Cluster.Network.SSL.Enabled {
		// TODO: proper SSL settings
		mc["ssl"] = map[string]any{
			"enabled":     true,
			"skip-verify": true,
		}
	}
	mc["log"] = map[string]any{
		"level": c.Logger.Level,
	}
	return mc
}
