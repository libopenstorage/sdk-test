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
	"testing"

	"github.com/libopenstorage/openstorage/api/server/sdk"
	"github.com/libopenstorage/openstorage/cluster"
	clustermanager "github.com/libopenstorage/openstorage/cluster/manager"
	osdconfig "github.com/libopenstorage/openstorage/config"
	"github.com/libopenstorage/openstorage/objectstore"
	"github.com/libopenstorage/openstorage/schedpolicy"
	volumedrivers "github.com/libopenstorage/openstorage/volume/drivers"

	"github.com/sirupsen/logrus"

	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/mem"
)

// This is used to test the SDK Sanity package
func TestSanity(t *testing.T) {

	// Initialize system
	kv, err := kvdb.New(mem.Name, "sdk_test", []string{}, nil, logrus.Panicf)
	if err != nil {
		t.Fatalf("Failed to initialize KVDB")
	}
	if err := kvdb.SetInstance(kv); err != nil {
		t.Fatalf("Failed to set KVDB instance")
	}
	// Initialize the cluster
	if err := clustermanager.Init(osdconfig.ClusterConfig{
		ClusterId:     "cluster",
		NodeId:        "1",
		DefaultDriver: "fake",
	}); err != nil {
		t.Fatalf("Unable to init cluster server: %v", err)
	}
	// Register the in-memory driver
	if err := volumedrivers.Register("fake", map[string]string{}); err != nil {
		t.Fatalf("Unable to start driver: %s", err)
	}
	cm, err := clustermanager.Inst()
	if err != nil {
		t.Fatalf("Unable to find cluster instance: %v", err)
	}
	go func() {
		if err := cm.StartWithConfiguration(
			0,
			false,
			"9002",
			&cluster.ClusterServerConfiguration{
				ConfigSchedManager:       schedpolicy.NewFakeScheduler(),
				ConfigObjectStoreManager: objectstore.NewfakeObjectstore(),
			},
		); err != nil {
			t.Fatalf("Unable to start cluster manager: %v", err)
		}
	}()

	// Start SDK Server
	sdkServer, err := sdk.New(&sdk.ServerConfig{
		Net:        "tcp",
		Address:    "127.0.0.1:0",
		DriverName: "fake",
		Cluster:    cm,
	})
	if err != nil {
		t.Fatalf("Failed to start SDK server for driver fake: %v", err)
	}
	sdkServer.Start()

	// Start OpenStorage SDK Stanity Tests
	Test(t, &SanityConfiguration{
		Address: sdkServer.Address(),
	})
}
