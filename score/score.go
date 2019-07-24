package score

import (
	"errors"
	"fmt"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/allegro/marathon-appcop/marathon"
	"github.com/allegro/marathon-appcop/metrics"
)

// Score contains score value to update, struct keeped inside Scorer as value as
// value as value as value
type Score struct {
	score      int
	lastUpdate time.Time
}

// Scorer keeps records of all applications behaviour on marathon
type Scorer struct {
	mutex            sync.RWMutex
	ScaleDownScore   int
	ResetInterval    time.Duration
	UpdateInterval   time.Duration
	EvaluateInterval time.Duration
	DryRun           bool
	ScaleLimit       int
	service          marathon.Marathoner
	scores           map[marathon.AppID]*Score
}

// Update struct for scoring specific app
type Update struct {
	// TODO(tz) to consider, store only AppID
	App    *marathon.App
	Update int
}

// New creates new scorer instance
func New(config Config, m marathon.Marathoner) (*Scorer, error) {

	if config.ResetInterval <= config.UpdateInterval {
		return nil, errors.New("UpdateInterval should be lower than ResetInterval")
	}

	if config.ResetInterval <= config.EvaluateInterval {
		return nil, errors.New("ResetInterval should be lower than EvaluateInterval")
	}

	return &Scorer{
		ScaleDownScore:   config.ScaleDownScore,
		ResetInterval:    config.ResetInterval,
		UpdateInterval:   config.UpdateInterval,
		EvaluateInterval: config.EvaluateInterval,
		ScaleLimit:       config.ScaleLimit,
		DryRun:           config.DryRun,
		service:          m,
		scores:           make(map[marathon.AppID]*Score),
	}, nil
}

// ScoreManager starts Scorer job
func (s *Scorer) ScoreManager() chan Update {
	updates := make(chan Update)

	log.Info("Starting ScoreManager")
	if s.DryRun {
		log.Info("DryRun, NOOP mode")
	}
	printTicker := time.NewTicker(s.UpdateInterval)
	evaluateTicker := time.NewTicker(s.EvaluateInterval)
	resetTimer := time.NewTicker(s.ResetInterval)

	go func() {
		for {
			select {
			case <-evaluateTicker.C:
				metrics.Mark("score.evaluates")
				go s.EvaluateApps()
			case <-printTicker.C:
				// Only used for debug purposes
				go s.printScores()
			case <-resetTimer.C:
				metrics.Mark("score.resets")
				go s.resetScores()
			case u := <-updates:
				metrics.UpdateGauge("score.updateQueue", int64(len(updates)))
				go s.initOrUpdateScore(u)
			}
		}
	}()
	return updates
}

func (s *Scorer) initOrUpdateScore(u Update) {
	log.WithFields(log.Fields{
		"appId":       u.App.ID,
		"scoreUpdate": u.Update,
	}).Debug("Score update")

	s.mutex.Lock()

	su := u.Update
	now := time.Now()

	if appScore, isScored := s.scores[u.App.ID]; isScored {
		appScore.score += su
		appScore.lastUpdate = now
	} else {
		s.scores[u.App.ID] = &Score{score: su, lastUpdate: now}
	}
	s.mutex.Unlock()
}

// if no such key, resetScores is noop
func (s *Scorer) resetScore(appID marathon.AppID) {
	s.mutex.Lock()

	delete(s.scores, appID)
	s.mutex.Unlock()
}

// Substracts score by configured treshold
// Noop if appID not exists
func (s *Scorer) subtractScore(appID marathon.AppID) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var ok bool
	var score *Score
	if score, ok = s.scores[appID]; !ok {
		return
	}

	score.score -= s.ScaleDownScore
}

func (s *Scorer) resetScores() {
	log.WithFields(log.Fields{
		"ScoresRecorded": len(s.scores),
	}).Debug("Reseting scores")

	s.mutex.Lock()

	s.scores = make(map[marathon.AppID]*Score)
	s.mutex.Unlock()
}

// EvaluateApps checks apps scores and if any is higher on score than limit,
// scale them down by one instance
func (s *Scorer) EvaluateApps() {

	i, err := s.evaluateApps()
	if err != nil && i == 0 {
		log.WithError(err).Error("Failed to evaluate")
	}
	log.Debugf("%d apps qualified for penalty", i)
}

func (s *Scorer) evaluateApps() (int, error) {
	limit := 2
	i := 0
	var lastErr error

	for appID, score := range s.scores {

		curScore := score.score
		// TODO(tz) - implement proper rate limiter with shared state accross goroutines
		// and configurable
		// https://gobyexample.com/rate-limiting
		if !(curScore > s.ScaleDownScore && i <= limit) {
			continue
		}

		err := s.scaleDown(appID)
		if err != nil {
			lastErr = err
			log.WithFields(log.Fields{
				"appId": appID,
				"Limit": i,
			}).Error(err)
			metrics.Mark("score.scale_fail")
			s.resetScore(appID)

			continue
		}

		metrics.Mark("score.scale_success")
		s.subtractScore(appID)

		i++
	}
	return i, lastErr
}

func (s *Scorer) scaleDown(appID marathon.AppID) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	log.WithFields(log.Fields{
		"appId": appID,
		"score": s.scores[appID].score,
	}).Info("Scaling down application")

	app, err := s.service.AppGet(appID)
	if err != nil {
		return err
	}

	// dry-run flag
	if s.DryRun {
		log.WithFields(log.Fields{
			"appId": appID,
			"score": s.scores[appID].score,
		}).Info("NOOP - App Scale Down")
		return nil
	}

	if app.HasImmunity() {
		// returning error up makes sure rate limiting works,
		// otherwise AppCop could loop over immune apps
		return fmt.Errorf("app: %s has immunity", app.ID)
	}

	err = s.service.AppScaleDown(app)
	return err

}

func (s *Scorer) printScores() {
	for app, score := range s.scores {
		log.WithFields(log.Fields{
			"app":   app,
			"score": score.score}).Debug("Output Scores")
	}
}
