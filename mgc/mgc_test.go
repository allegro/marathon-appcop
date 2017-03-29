package mgc

import (
	"testing"
	"time"

	"github.com/allegro/marathon-appcop/marathon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMGCConstructorShouldReturnStruct(t *testing.T) {
	t.Parallel()
	// given
	marathon := marathon.Marathon{}
	config := Config{}
	timeNow := time.Now()
	given := MarathonGC{config: config,
		marathon:    marathon,
		apps:        nil,
		lastRefresh: timeNow,
	}

	// when
	mgc, err := New(config, marathon)

	// then
	require.NoError(t, err)
	assert.ObjectsAreEqual(given, mgc)

}

func TestMGCConstructorShouldReturnDifferentStruct(t *testing.T) {
	t.Parallel()
	// given
	marathon := marathon.Marathon{}
	config := Config{}
	timeNow := time.Now()
	given := MarathonGC{config: config,
		marathon:    nil,
		apps:        nil,
		lastRefresh: timeNow,
	}

	// when
	mgc, err := New(config, marathon)

	// then
	require.NoError(t, err)
	assert.NotEqual(t, given, mgc)

}

func TestAppCoppedApplicationRetrunsTrue(t *testing.T) {
	t.Parallel()
	// given
	labels := map[string]string{
		"appcop": "suspend",
		"consul": "true",
	}
	app := &marathon.App{
		ID:     "/test/app",
		Labels: labels,
	}

	// when
	appcoped := appCopped(app)
	//then
	assert.True(t, appcoped)
}

func TestAppCoppedApplicationRetrunsFalse(t *testing.T) {
	t.Parallel()
	// given
	labels := map[string]string{
		"consul": "true",
	}
	app := &marathon.App{
		ID:     "/test/app",
		Labels: labels,
	}

	// when
	appcoped := appCopped(app)
	//then
	assert.True(t, !appcoped)
}

func TestAppCoppedApplicationNilLabels(t *testing.T) {
	t.Parallel()
	// given
	app := &marathon.App{
		ID: "/test/app",
	}
	// when
	appcoped := appCopped(app)
	//then
	assert.True(t, !appcoped)
}

func TestGetOldSuspendedWhenMarathonReturnsOneOldSuspendedApp(t *testing.T) {
	t.Parallel()
	//given
	m := marathon.Marathon{}
	config := Config{}
	given, _ := New(config, m)
	wayBack := "2006-01-02T15:04:05.000Z"
	given.apps = []*marathon.App{
		{VersionInfo: marathon.VersionInfo{
			LastScalingAt:      wayBack,
			LastConfigChangeAt: wayBack},
		},
	}
	// when
	apps := given.getOldSuspended()
	// then
	assert.Equal(t, 1, len(apps))
	assert.NotNil(t, apps)

}

func TestGetOldSuspendedWhenMarathonReturnsTwoOldSuspendedApps(t *testing.T) {
	t.Parallel()
	//given
	m := marathon.Marathon{}
	config := Config{}
	given, _ := New(config, m)
	wayBack := "2006-01-02T15:04:05.000Z"
	alsoWayBack := "2007-01-02T15:04:05.000Z"
	given.apps = []*marathon.App{
		{VersionInfo: marathon.VersionInfo{
			LastScalingAt:      wayBack,
			LastConfigChangeAt: wayBack},
		},
		{VersionInfo: marathon.VersionInfo{
			LastScalingAt:      alsoWayBack,
			LastConfigChangeAt: alsoWayBack},
		},
	}
	// when
	apps := given.getOldSuspended()
	// then
	assert.Equal(t, 2, len(apps))
	assert.NotNil(t, apps)

}

func TestGetOneOldSuspendedWhenMarathonReturnsTwoOldSuspendedApps(t *testing.T) {
	t.Parallel()
	//given
	m := marathon.Marathon{}
	config := Config{}
	given, _ := New(config, m)
	ti := time.Now()
	timeNow := ti.Format("2006-01-02T15:04:05.000Z")

	wayBack := "2006-01-02T15:04:05.000Z"
	given.apps = []*marathon.App{
		{VersionInfo: marathon.VersionInfo{
			LastScalingAt:      wayBack,
			LastConfigChangeAt: wayBack},
		},
		{VersionInfo: marathon.VersionInfo{
			LastScalingAt:      timeNow,
			LastConfigChangeAt: timeNow},
		},
	}
	given.config.MaxSuspendTime = 1300000000000
	// when
	apps := given.getOldSuspended()
	// then
	assert.Equal(t, 1, len(apps))
	assert.NotNil(t, apps)

}

func TestGetOneOldSuspendedWhenMGCIsConfiguredToSuspendOnlyAppCopped(t *testing.T) {
	t.Parallel()
	//given
	m := marathon.Marathon{}
	config := Config{AppCopOnly: true}
	given, _ := New(config, m)
	wayBack := "2006-01-02T15:04:05.000Z"
	given.apps = []*marathon.App{
		{VersionInfo: marathon.VersionInfo{
			LastScalingAt:      wayBack,
			LastConfigChangeAt: wayBack},
		},
		{
			VersionInfo: marathon.VersionInfo{
				LastScalingAt:      wayBack,
				LastConfigChangeAt: wayBack},
			Labels: map[string]string{"appcop": "suspended"},
		},
	}
	// when
	apps := given.getOldSuspended()
	// then
	assert.Equal(t, 1, len(apps))
	assert.NotNil(t, apps)
}

func TestGetOldSuspendedReturnsNothingWhenMGCIsConfiguredToSuspendOnlyAppCopped(t *testing.T) {
	t.Parallel()
	//given
	m := marathon.Marathon{}
	config := Config{AppCopOnly: true}
	given, _ := New(config, m)
	wayBack := "2006-01-02T15:04:05.000Z"
	given.apps = []*marathon.App{
		{VersionInfo: marathon.VersionInfo{
			LastScalingAt:      wayBack,
			LastConfigChangeAt: wayBack},
		},
		{VersionInfo: marathon.VersionInfo{
			LastScalingAt:      wayBack,
			LastConfigChangeAt: wayBack},
		},
	}
	// when
	apps := given.getOldSuspended()
	// then
	assert.Equal(t, 0, len(apps))
	assert.Nil(t, apps)

}

func TestGCAbleReturnsFalseWhenInstanceNumIsGreaterThanZero(t *testing.T) {
	t.Parallel()
	//given
	m := marathon.Marathon{}
	mgc, err := New(Config{}, m)
	app := &marathon.App{Instances: 0}
	// when
	able := mgc.shouldBeCollected(app)
	// then
	require.NoError(t, err)
	assert.False(t, able)
}

func TestGCAbleReturnsTrue(t *testing.T) {
	t.Parallel()
	//given
	m := marathon.Marathon{}
	mgc, err := New(Config{}, m)
	wayBack := "2006-01-02T15:04:05.000Z"
	app := &marathon.App{
		VersionInfo: marathon.VersionInfo{
			LastScalingAt:      wayBack,
			LastConfigChangeAt: wayBack},
		Instances: 0,
	}
	// when
	able := mgc.shouldBeCollected(app)
	// then
	require.NoError(t, err)
	assert.True(t, able)
}

func TestGCAbleReturnsFalseWhenParsingErrorOccures(t *testing.T) {
	t.Parallel()
	//given
	m := marathon.Marathon{}
	mgc, _ := New(Config{}, m)
	wayBack := "200aaa6-01-02T15:04:05.000Z"
	app := &marathon.App{
		VersionInfo: marathon.VersionInfo{
			LastScalingAt:      wayBack,
			LastConfigChangeAt: wayBack},
		Instances: 0,
	}
	// when
	able := mgc.shouldBeCollected(app)
	// then
	assert.False(t, able)
}

func TestMGCRefreshSuccessWhenMarathonReturnsTwoApps(t *testing.T) {
	t.Parallel()

	// given
	apps := []*marathon.App{
		{ID: "firstApp", Instances: 1},
		{ID: "secondApp", Instances: 2},
	}
	m := marathon.MStub{Apps: apps}
	mgc, _ := New(Config{}, m)

	// when
	err := mgc.refresh()

	// then
	require.NoError(t, err)
	assert.NotNil(t, mgc)
	assert.Equal(t, mgc.apps, apps)
}

func TestMGCRefreshErrorWhenMarathonGetAppsReturnsError(t *testing.T) {
	t.Parallel()
	// given
	apps := []*marathon.App{
		{ID: "firstApp", Instances: 1},
		{ID: "secondApp", Instances: 2},
	}
	m := marathon.MStub{Apps: apps, AppsGetFail: true}
	mgc, _ := New(Config{}, m)
	// when
	err := mgc.refresh()
	// then
	require.Error(t, err)
	assert.NotNil(t, mgc)
	assert.Equal(t, []*marathon.App(nil), mgc.apps)
}

func TestMGCGroupDeleteWhenMrathonReturnsSuccess(t *testing.T) {
	t.Parallel()
	// given
	m := marathon.MStub{}
	mgc, _ := New(Config{}, m)
	// when
	err := mgc.groupDelete("testgroup")
	//then
	require.NoError(t, err)
}

func TestMGCGroupDeleteWhenMrathonReturnsError(t *testing.T) {
	t.Parallel()
	// given
	m := marathon.MStub{GroupDelFail: true}
	mgc, _ := New(Config{}, m)
	// when
	err := mgc.groupDelete("testgroup")
	//then
	require.Error(t, err)
}

func TestMGCDeleteSuspendedAppsWhenThereIsOneAppToDelete(t *testing.T) {
	t.Parallel()
	// given
	apps := []*marathon.App{
		{ID: "testapp0"},
	}
	m := marathon.MStub{Apps: apps}
	mgc, _ := New(Config{}, m)
	// when
	i := mgc.deleteSuspended(apps)
	// then
	assert.Equal(t, 1, i)
}

func TestMGCDeleteSuspendedAppsWhenThereIsTwoAppsToDelete(t *testing.T) {
	t.Parallel()
	// given
	apps := []*marathon.App{
		{ID: "testapp0"},
		{ID: "testapp1"},
	}
	m := marathon.MStub{Apps: apps}
	mgc, _ := New(Config{}, m)
	// when
	i := mgc.deleteSuspended(apps)
	// then
	assert.Equal(t, 2, i)
}

func TestMGCDeleteSuspendedAppsWhenMarathonReturnsErrorsOnSomeDeletes(t *testing.T) {
	t.Parallel()
	// given
	apps := []*marathon.App{
		{ID: "testapp0"},
		{ID: "testapp1"},
		{ID: "testapp2"},
		{ID: "testapp3"},
	}
	failCounter := &marathon.FailCounter{Counter: 1}
	m := marathon.MStub{Apps: apps, AppDelHalfFail: true, FailCounter: failCounter}
	mgc, _ := New(Config{}, m)
	// when
	i := mgc.deleteSuspended(apps)
	// then
	assert.Equal(t, 2, i)
}
