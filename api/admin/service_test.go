// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package admin

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"go.uber.org/mock/gomock"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/avalanchego/vms"
	"github.com/ava-labs/avalanchego/vms/registry"
)

type loadVMsTest struct {
	admin          *Admin
	ctrl           *gomock.Controller
	mockVMManager  *vms.MockManager
	mockVMRegistry *registry.MockVMRegistry
}

func initLoadVMsTest(t *testing.T) *loadVMsTest {
	ctrl := gomock.NewController(t)

	mockVMRegistry := registry.NewMockVMRegistry(ctrl)
	mockVMManager := vms.NewMockManager(ctrl)

	return &loadVMsTest{
		admin: &Admin{Config: Config{
			Log:        logging.NoLog{},
			VMRegistry: mockVMRegistry,
			VMManager:  mockVMManager,
		}},
		ctrl:           ctrl,
		mockVMManager:  mockVMManager,
		mockVMRegistry: mockVMRegistry,
	}
}

// Tests behavior for LoadVMs if everything succeeds.
func TestLoadVMsSuccess(t *testing.T) {
	require := require.New(t)

	resources := initLoadVMsTest(t)

	id1 := ids.GenerateTestID()
	id2 := ids.GenerateTestID()

	newVMs := []ids.ID{id1, id2}
	failedVMs := map[ids.ID]error{
		ids.GenerateTestID(): errTest,
	}
	// every vm is at least aliased to itself.
	alias1 := []string{id1.String(), "vm1-alias-1", "vm1-alias-2"}
	alias2 := []string{id2.String(), "vm2-alias-1", "vm2-alias-2"}
	// we expect that we dedup the redundant alias of vmId.
	expectedVMRegistry := map[ids.ID][]string{
		id1: alias1[1:],
		id2: alias2[1:],
	}

	resources.mockVMRegistry.EXPECT().Reload(gomock.Any()).Times(1).Return(newVMs, failedVMs, nil)
	resources.mockVMManager.EXPECT().Aliases(id1).Times(1).Return(alias1, nil)
	resources.mockVMManager.EXPECT().Aliases(id2).Times(1).Return(alias2, nil)

	// execute test
	reply := LoadVMsReply{}
	require.NoError(resources.admin.LoadVMs(&http.Request{}, nil, &reply))
	require.Equal(expectedVMRegistry, reply.NewVMs)
}

// Tests behavior for LoadVMs if we fail to reload vms.
func TestLoadVMsReloadFails(t *testing.T) {
	require := require.New(t)

	resources := initLoadVMsTest(t)

	// Reload fails
	resources.mockVMRegistry.EXPECT().Reload(gomock.Any()).Times(1).Return(nil, nil, errTest)

	reply := LoadVMsReply{}
	err := resources.admin.LoadVMs(&http.Request{}, nil, &reply)
	require.ErrorIs(err, errTest)
}

// Tests behavior for LoadVMs if we fail to fetch our aliases
func TestLoadVMsGetAliasesFails(t *testing.T) {
	require := require.New(t)

	resources := initLoadVMsTest(t)

	id1 := ids.GenerateTestID()
	id2 := ids.GenerateTestID()
	newVMs := []ids.ID{id1, id2}
	failedVMs := map[ids.ID]error{
		ids.GenerateTestID(): errTest,
	}
	// every vm is at least aliased to itself.
	alias1 := []string{id1.String(), "vm1-alias-1", "vm1-alias-2"}

	resources.mockVMRegistry.EXPECT().Reload(gomock.Any()).Times(1).Return(newVMs, failedVMs, nil)
	resources.mockVMManager.EXPECT().Aliases(id1).Times(1).Return(alias1, nil)
	resources.mockVMManager.EXPECT().Aliases(id2).Times(1).Return(nil, errTest)

	reply := LoadVMsReply{}
	err := resources.admin.LoadVMs(&http.Request{}, nil, &reply)
	require.ErrorIs(err, errTest)
}
