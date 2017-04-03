package score

import (
	"testing"
	"time"

	"github.com/allegro/marathon-appcop/marathon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestScorer() (*Scorer, error) {
	return New(Config{false, 1, 1, 3, 2, 1}, nil)
}

func TestNewProvidedConfigContainsUnsensibleValuesReturnsErrorAndNilScorer(t *testing.T) {
	t.Parallel()
	// given
	c := Config{
		ScaleDownScore:   1,
		UpdateInterval:   1,
		ResetInterval:    1,
		EvaluateInterval: 1,
		ScaleLimit:       1,
	}
	// when
	scorer, err := New(c, nil)
	//then
	assert.Error(t, err)
	assert.Equal(t, scorer, (*Scorer)(nil))
}

func TestNewReturnsCorrectlyInitializedScorerAndNoError(t *testing.T) {
	t.Parallel()
	// given
	c := Config{
		ScaleDownScore:   1,
		UpdateInterval:   1,
		ResetInterval:    3,
		EvaluateInterval: 2,
		ScaleLimit:       1,
		DryRun:           false,
	}
	// when
	expectedScorer := &Scorer{
		ScaleDownScore:   1,
		ResetInterval:    3,
		UpdateInterval:   1,
		EvaluateInterval: 2,
		ScaleLimit:       1,
		scores:           map[marathon.AppID]*Score{},
	}
	actualScorer, err := New(c, nil)
	//then
	assert.Equal(t, expectedScorer, actualScorer)
	assert.Nil(t, err)
}

var initOrUpdateTestCases = []struct {
	updates        []Update
	expectedScores map[marathon.AppID]*Score
}{
	{
		updates: []Update{
			{App: &marathon.App{ID: "appid"}, Update: 1},
		},
		expectedScores: map[marathon.AppID]*Score{
			marathon.AppID("appid"): {1, time.Now()},
		},
	},
	{
		updates: []Update{
			{App: &marathon.App{ID: "appid"}, Update: 1},
			{App: &marathon.App{ID: "appid"}, Update: 1},
			{App: &marathon.App{ID: "appid"}, Update: 1},
			{App: &marathon.App{ID: "appid"}, Update: 1},
		},
		expectedScores: map[marathon.AppID]*Score{
			marathon.AppID("appid"): {4, time.Now()},
		},
	},
	{
		updates: []Update{
			{App: &marathon.App{ID: "appid0"}, Update: 1},
			{App: &marathon.App{ID: "appid0"}, Update: 1},
			{App: &marathon.App{ID: "appid1"}, Update: 1},
			{App: &marathon.App{ID: "appid1"}, Update: 1},
		},
		expectedScores: map[marathon.AppID]*Score{
			marathon.AppID("appid0"): {2, time.Now()},
			marathon.AppID("appid1"): {2, time.Now()},
		},
	},
	{
		updates: []Update{
			{App: &marathon.App{ID: "appid0"}, Update: -1},
			{App: &marathon.App{ID: "appid0"}, Update: 1},
			{App: &marathon.App{ID: "appid1"}, Update: -1},
			{App: &marathon.App{ID: "appid1"}, Update: -1},
		},
		expectedScores: map[marathon.AppID]*Score{
			marathon.AppID("appid0"): {0, time.Now()},
			marathon.AppID("appid1"): {-2, time.Now()},
		},
	},
	{
		updates: []Update{
			{App: &marathon.App{ID: "appid0"}, Update: -1},
			{App: &marathon.App{ID: "appid1"}, Update: 1},
			{App: &marathon.App{ID: "appid2"}, Update: -1},
			{App: &marathon.App{ID: "appid3"}, Update: -1},
		},
		expectedScores: map[marathon.AppID]*Score{
			marathon.AppID("appid0"): {-1, time.Now()},
			marathon.AppID("appid1"): {1, time.Now()},
			marathon.AppID("appid2"): {-1, time.Now()},
			marathon.AppID("appid3"): {-1, time.Now()},
		},
	},
}

func TestInitOrUpdateTestCases(t *testing.T) {
	t.Parallel()
	for _, testCase := range initOrUpdateTestCases {
		s, err := newTestScorer()
		require.NoError(t, err)
		for _, update := range testCase.updates {
			s.initOrUpdateScore(update)
		}
		// check assertions
		for appID, score := range testCase.expectedScores {
			app, ok := s.scores[appID]
			require.True(t, ok)
			assert.Equal(t, score.score, app.score)
		}
	}
}

func TestResetScoreDeletesSpecifiedAppFromScorerLeaveRestUntoutchCheckScoreMap(t *testing.T) {
	t.Parallel()
	// given
	s, err := newTestScorer()
	require.NoError(t, err)
	s.scores["testapp0"] = &Score{score: 1, lastUpdate: time.Now()}
	// Shouldnt be touch
	s.scores["testapp1"] = &Score{score: 2, lastUpdate: time.Now()}
	expectedScore := 2
	// when
	s.resetScore("testapp0")
	_, ok0 := s.scores["testapp0"]
	app, ok1 := s.scores["testapp1"]
	//then
	assert.False(t, ok0)
	assert.True(t, ok1)
	assert.Equal(t, expectedScore, app.score)
}

var substractScoreTestCases = []struct {
	initialScores            map[marathon.AppID]int
	appsToSubstractScoreFrom []marathon.AppID
	expectedScores           map[marathon.AppID]int
}{
	{
		initialScores:            map[marathon.AppID]int{},
		appsToSubstractScoreFrom: []marathon.AppID{},
		expectedScores:           map[marathon.AppID]int{},
	},
	{
		initialScores: map[marathon.AppID]int{
			marathon.AppID("id1"): 1,
			marathon.AppID("id2"): 2,
		},
		appsToSubstractScoreFrom: []marathon.AppID{
			marathon.AppID("id1"), marathon.AppID("id2"),
		},
		expectedScores: map[marathon.AppID]int{
			marathon.AppID("id1"): 0,
			marathon.AppID("id2"): 1,
		},
	},
	{
		initialScores: map[marathon.AppID]int{
			marathon.AppID("id1"): 20,
			marathon.AppID("id2"): 30,
		},
		appsToSubstractScoreFrom: []marathon.AppID{
			marathon.AppID("id1"), marathon.AppID("id2"),
		},
		expectedScores: map[marathon.AppID]int{
			marathon.AppID("id1"): 19,
			marathon.AppID("id2"): 29,
		},
	},
	{
		initialScores: map[marathon.AppID]int{
			marathon.AppID("id1"): -1,
			marathon.AppID("id2"): -2,
		},
		appsToSubstractScoreFrom: []marathon.AppID{
			marathon.AppID("id1"), marathon.AppID("id2"),
		},
		expectedScores: map[marathon.AppID]int{
			marathon.AppID("id1"): -2,
			marathon.AppID("id2"): -3,
		},
	},
}

func TestSubstractScoresTestCases(t *testing.T) {
	t.Parallel()
	for _, testCase := range substractScoreTestCases {
		scorer, err := newTestScorer()
		require.NoError(t, err)
		// feed scores
		for app, score := range testCase.initialScores {
			scorer.scores[app] = &Score{score, time.Now()}
		}
		// actual substraction
		for _, app := range testCase.appsToSubstractScoreFrom {
			scorer.subtractScore(app)
		}
		// check assertions
		for app, expectedScore := range testCase.expectedScores {
			appScore := scorer.scores[app].score
			assert.Equal(t, expectedScore, appScore)
		}
	}
}

var evaluateScoresTestCases = []struct {
	scaleDownScore       int
	initialScores        map[marathon.AppID]int
	expectedAppsToPacify int
}{
	{
		scaleDownScore: 20,
		initialScores: map[marathon.AppID]int{
			marathon.AppID("id1"): 1,
			marathon.AppID("id2"): 2,
		},
		expectedAppsToPacify: 0,
	},
	{
		scaleDownScore: 20,
		initialScores: map[marathon.AppID]int{
			marathon.AppID("id1"): 21,
			marathon.AppID("id2"): 3,
		},
		expectedAppsToPacify: 1,
	},
	{
		scaleDownScore: 20,
		initialScores: map[marathon.AppID]int{
			marathon.AppID("id1"): 21,
			marathon.AppID("id2"): 3,
			marathon.AppID("id3"): -1,
		},
		expectedAppsToPacify: 1,
	},
	{
		scaleDownScore: 20,
		initialScores: map[marathon.AppID]int{
			marathon.AppID("id1"): 1230,
			marathon.AppID("id2"): 3,
			marathon.AppID("id3"): -1,
		},
		expectedAppsToPacify: 1,
	},
}

func TestEvaluateScoresTestCases(t *testing.T) {
	t.Parallel()
	for _, testCase := range evaluateScoresTestCases {
		scaleCounter := &marathon.ScaleCounter{Counter: 0}
		m := marathon.MStub{ScaleCounter: scaleCounter}
		scorer, err := New(Config{false, testCase.scaleDownScore, 1, 3, 2, 1}, m)
		require.NoError(t, err)
		// feed scores
		for app, score := range testCase.initialScores {
			scorer.scores[app] = &Score{score, time.Now()}
		}
		// actual evaluation
		appsToPacify, _ := scorer.evaluateApps()
		// check assertions
		assert.Equal(t, testCase.expectedAppsToPacify, appsToPacify)
		// expectedAppsToPacify equals ScaleCounter increments
		assert.Equal(t, testCase.expectedAppsToPacify, m.ScaleCounter.Counter)
	}
}

func TestEvaluateScoresTestCasesWithDryRunTrue(t *testing.T) {
	t.Parallel()
	for _, testCase := range evaluateScoresTestCases {
		scaleCounter := &marathon.ScaleCounter{Counter: 0}
		m := marathon.MStub{ScaleCounter: scaleCounter}
		scorer, err := New(Config{true, testCase.scaleDownScore, 1, 3, 2, 1}, m)
		require.NoError(t, err)
		// feed scores
		for app, score := range testCase.initialScores {
			scorer.scores[app] = &Score{score, time.Now()}
		}
		// actual evaluation
		appsToPacify, _ := scorer.evaluateApps()
		// check assertions
		// check how many apps are above threshold
		assert.Equal(t, testCase.expectedAppsToPacify, appsToPacify)
		// expectedAppsToPacify equals ScaleCounter increments
		// when dry run -> never pacify
		assert.Equal(t, 0, m.ScaleCounter.Counter)
	}
}
