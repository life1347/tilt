package engine

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/windmilleng/tilt/internal/k8s"
	"github.com/windmilleng/tilt/internal/model"
	"github.com/windmilleng/tilt/internal/store"
	"github.com/windmilleng/tilt/internal/testutils/tempdir"
	"k8s.io/api/core/v1"
)

func TestPortForward(t *testing.T) {
	f := newPLCFixture(t)
	defer f.TearDown()

	state := f.st.LockMutableState()
	state.ManifestStates["fe"] = &store.ManifestState{
		Manifest: model.Manifest{
			Name: "fe",
			PortForwards: []model.PortForward{
				{
					LocalPort:     8080,
					ContainerPort: 8081,
				},
			},
		},
	}
	f.st.UnlockMutableState()

	f.plc.OnChange(f.ctx, f.st)
	assert.Equal(t, 0, len(f.plc.activeForwards))

	state = f.st.LockMutableState()
	state.ManifestStates["fe"].Pod = store.Pod{PodID: "pod-id", Phase: v1.PodRunning}
	f.st.UnlockMutableState()

	f.plc.OnChange(f.ctx, f.st)
	assert.Equal(t, 1, len(f.plc.activeForwards))
	assert.Equal(t, "pod-id", f.kCli.LastForwardPortPodID.String())

	state = f.st.LockMutableState()
	state.ManifestStates["fe"].Pod = store.Pod{PodID: "pod-id2", Phase: v1.PodRunning}
	f.st.UnlockMutableState()

	f.plc.OnChange(f.ctx, f.st)
	assert.Equal(t, 1, len(f.plc.activeForwards))
	assert.Equal(t, "pod-id2", f.kCli.LastForwardPortPodID.String())

	state = f.st.LockMutableState()
	state.ManifestStates["fe"].Pod = store.Pod{PodID: "pod-id2", Phase: v1.PodPending}
	f.st.UnlockMutableState()

	f.plc.OnChange(f.ctx, f.st)
	assert.Equal(t, 0, len(f.plc.activeForwards))
}

type plcFixture struct {
	*tempdir.TempDirFixture
	ctx  context.Context
	kCli *k8s.FakeK8sClient
	st   *store.Store
	plc  *PortForwardController
}

func newPLCFixture(t *testing.T) *plcFixture {
	f := tempdir.NewTempDirFixture(t)
	st := store.NewStore()
	kCli := k8s.NewFakeK8sClient()
	plc := NewPortForwardController(kCli)
	return &plcFixture{
		TempDirFixture: f,
		ctx:            context.Background(),
		st:             st,
		kCli:           kCli,
		plc:            plc,
	}
}