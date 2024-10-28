package root

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	corestore "cosmossdk.io/core/store"
	coretesting "cosmossdk.io/core/testing"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/metrics"
	"cosmossdk.io/store/v2/mock"
	"cosmossdk.io/store/v2/proof"
	"cosmossdk.io/store/v2/pruning"
)

func newTestRootStore(ss store.VersionedWriter, sc store.Committer) *Store {
	noopLog := coretesting.NewNopLogger()
	pm := pruning.NewManager(sc.(store.Pruner), ss.(store.Pruner), nil, nil)
	return &Store{
		logger:          noopLog,
		telemetry:       metrics.Metrics{},
		initialVersion:  1,
		stateStorage:    ss,
		stateCommitment: sc,
		pruningManager:  pm,
		isMigrating:     false,
	}
}

func TestGetLatestState(t *testing.T) {
	ctrl := gomock.NewController(t)
	ss := mock.NewMockStateStorage(ctrl)
	sc := mock.NewMockStateCommitter(ctrl)
	rs := newTestRootStore(ss, sc)

	// Get the latest version
	sc.EXPECT().GetLatestVersion().Return(uint64(0), errors.New("error"))
	_, err := rs.GetLatestVersion()
	require.Error(t, err)
	sc.EXPECT().GetLatestVersion().Return(uint64(1), nil)
	v, err := rs.GetLatestVersion()
	require.NoError(t, err)
	require.Equal(t, uint64(1), v)
}

func TestQuery(t *testing.T) {
	ctrl := gomock.NewController(t)
	ss := mock.NewMockStateStorage(ctrl)
	sc := mock.NewMockStateCommitter(ctrl)
	rs := newTestRootStore(ss, sc)

	// Query without Proof
	ss.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))
	_, err := rs.Query(nil, 0, nil, false)
	require.Error(t, err)
	ss.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)
	sc.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))
	_, err = rs.Query(nil, 0, nil, false)
	require.Error(t, err)
	ss.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)
	sc.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte("value"), nil)
	v, err := rs.Query(nil, 0, nil, false)
	require.NoError(t, err)
	require.Equal(t, []byte("value"), v.Value)
	ss.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte("value"), nil)
	v, err = rs.Query(nil, 0, nil, false)
	require.NoError(t, err)
	require.Equal(t, []byte("value"), v.Value)

	// Query with Proof
	ss.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte("value"), nil)
	sc.EXPECT().GetProof(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))
	v, err = rs.Query(nil, 0, nil, true)
	require.Error(t, err)

	// Query with Migration
	rs.isMigrating = true
	sc.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))
	_, err = rs.Query(nil, 0, nil, false)
	require.Error(t, err)
}

func TestLoadVersion(t *testing.T) {
	ctrl := gomock.NewController(t)
	ss := mock.NewMockStateStorage(ctrl)
	sc := mock.NewMockStateCommitter(ctrl)
	rs := newTestRootStore(ss, sc)

	// LoadLatestVersion
	sc.EXPECT().GetLatestVersion().Return(uint64(0), errors.New("error"))
	err := rs.LoadLatestVersion()
	require.Error(t, err)
	sc.EXPECT().GetLatestVersion().Return(uint64(1), nil)
	sc.EXPECT().LoadVersion(uint64(1)).Return(errors.New("error"))
	err = rs.LoadLatestVersion()
	require.Error(t, err)

	// LoadVersion
	sc.EXPECT().LoadVersion(gomock.Any()).Return(nil)
	sc.EXPECT().GetCommitInfo(uint64(2)).Return(nil, errors.New("error"))
	err = rs.LoadVersion(uint64(2))
	require.Error(t, err)

	// LoadVersionUpgrade
	v := &corestore.StoreUpgrades{}
	sc.EXPECT().LoadVersionAndUpgrade(uint64(2), v).Return(errors.New("error"))
	err = rs.LoadVersionAndUpgrade(uint64(2), v)
	require.Error(t, err)
	sc.EXPECT().LoadVersionAndUpgrade(uint64(2), v).Return(nil)
	sc.EXPECT().GetCommitInfo(uint64(2)).Return(nil, nil)
	ss.EXPECT().PruneStoreKeys(gomock.Any(), uint64(2)).Return(errors.New("error"))
	err = rs.LoadVersionAndUpgrade(uint64(2), v)
	require.Error(t, err)

	// LoadVersionUpgrade with Migration
	rs.isMigrating = true
	err = rs.LoadVersionAndUpgrade(uint64(2), v)
	require.Error(t, err)
}

func TestWorkingHahs(t *testing.T) {
	ctrl := gomock.NewController(t)
	ss := mock.NewMockStateStorage(ctrl)
	sc := mock.NewMockStateCommitter(ctrl)
	rs := newTestRootStore(ss, sc)

	cs := corestore.NewChangeset()
	// writeSC test
	sc.EXPECT().WriteChangeset(cs).Return(errors.New("error"))
	err := rs.writeSC(cs)
	require.Error(t, err)
	sc.EXPECT().WriteChangeset(cs).Return(nil)
	sc.EXPECT().WorkingCommitInfo(gomock.Any()).Return(nil)
	err = rs.writeSC(cs)
	require.NoError(t, err)

	// WorkingHash test
	sc.EXPECT().WriteChangeset(cs).Return(nil)
	sc.EXPECT().WorkingCommitInfo(gomock.Any()).Return(nil)
	ss.EXPECT().ApplyChangeset(gomock.Any(), cs).Return(errors.New("error"))
	_, err = rs.WorkingHash(cs)
	require.Error(t, err)
	sc.EXPECT().WriteChangeset(cs).Return(nil)
	sc.EXPECT().WorkingCommitInfo(gomock.Any()).Return(nil)
	ss.EXPECT().ApplyChangeset(gomock.Any(), cs).Return(errors.New("error"))
	_, err = rs.WorkingHash(cs)
	require.Error(t, err)
	sc.EXPECT().WriteChangeset(cs).Return(nil)
	sc.EXPECT().WorkingCommitInfo(gomock.Any()).Return(&proof.CommitInfo{})
	ss.EXPECT().ApplyChangeset(gomock.Any(), cs).Return(nil)
	_, err = rs.WorkingHash(cs)
	require.NoError(t, err)
}

func TestCommit(t *testing.T) {
	ctrl := gomock.NewController(t)
	ss := mock.NewMockStateStorage(ctrl)
	sc := mock.NewMockStateCommitter(ctrl)
	rs := newTestRootStore(ss, sc)

	cs := corestore.NewChangeset()
	// test commitSC
	rs.lastCommitInfo = &proof.CommitInfo{}
	sc.EXPECT().Commit(gomock.Any()).Return(nil, errors.New("error"))
	err := rs.commitSC()
	require.Error(t, err)
	sc.EXPECT().Commit(gomock.Any()).Return(&proof.CommitInfo{CommitHash: []byte("wrong hash"), StoreInfos: []proof.StoreInfo{{}}}, nil) // wrong hash
	err = rs.commitSC()
	require.Error(t, err)

	// Commit test
	sc.EXPECT().WriteChangeset(cs).Return(errors.New("error"))
	_, err = rs.Commit(cs)
	require.Error(t, err)
	sc.EXPECT().WriteChangeset(cs).Return(nil)
	sc.EXPECT().WorkingCommitInfo(gomock.Any()).Return(&proof.CommitInfo{})
	sc.EXPECT().PausePruning(gomock.Any()).Return()
	ss.EXPECT().PausePruning(gomock.Any()).Return()
	ss.EXPECT().ApplyChangeset(gomock.Any(), cs).Return(nil)
	sc.EXPECT().Commit(gomock.Any()).Return(nil, errors.New("error"))
	_, err = rs.Commit(cs)
	require.Error(t, err)
	sc.EXPECT().WriteChangeset(cs).Return(nil)
	sc.EXPECT().WorkingCommitInfo(gomock.Any()).Return(&proof.CommitInfo{})
	sc.EXPECT().PausePruning(gomock.Any()).Return()
	ss.EXPECT().PausePruning(gomock.Any()).Return()
	ss.EXPECT().ApplyChangeset(gomock.Any(), cs).Return(errors.New("error"))
	sc.EXPECT().Commit(gomock.Any()).Return(&proof.CommitInfo{}, nil)
	_, err = rs.Commit(cs)
	require.Error(t, err)
	sc.EXPECT().WriteChangeset(cs).Return(nil)
	sc.EXPECT().WorkingCommitInfo(gomock.Any()).Return(&proof.CommitInfo{})
	sc.EXPECT().PausePruning(true).Return()
	ss.EXPECT().PausePruning(true).Return()
	ss.EXPECT().ApplyChangeset(gomock.Any(), cs).Return(nil)
	sc.EXPECT().Commit(gomock.Any()).Return(&proof.CommitInfo{}, nil)
	sc.EXPECT().PausePruning(false).Return()
	ss.EXPECT().PausePruning(false).Return()
	_, err = rs.Commit(cs)
	require.NoError(t, err)
}
