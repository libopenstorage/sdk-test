/*
Copyright 2018 Portworx

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sanity

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	config *SanityConfiguration
	conn   *grpc.ClientConn
	lock   sync.Mutex
)

type SanityConfiguration struct {
	Address string
}

// Test will test start the sanity tests
func Test(t *testing.T, reqConfig *SanityConfiguration) {
	lock.Lock()
	defer lock.Unlock()

	config = reqConfig
	RegisterFailHandler(Fail)
	RunSpecs(t, "OpenStorage SDK Test Suite")
}

var _ = BeforeSuite(func() {
	var err error

	By("connecting to OpenStorage SDK endpoint")
	conn, err = connect(config.Address)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	conn.Close()
})

// Connect address by grpc
func connect(address string) (*grpc.ClientConn, error) {
	dialOptions := []grpc.DialOption{
		grpc.WithInsecure(),
	}
	u, err := url.Parse(address)
	if err == nil && (!u.IsAbs() || u.Scheme == "unix") {
		dialOptions = append(dialOptions,
			grpc.WithDialer(
				func(addr string, timeout time.Duration) (net.Conn, error) {
					return net.DialTimeout("unix", u.Path, timeout)
				}))
	}

	conn, err := grpc.Dial(address, dialOptions...)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	for {
		if !conn.WaitForStateChange(ctx, conn.GetState()) {
			return conn, fmt.Errorf("Connection timed out")
		}
		if conn.GetState() == connectivity.Ready {
			return conn, nil
		}
	}
}
