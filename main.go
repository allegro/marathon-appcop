package main

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/allegro/marathon-appcop/config"
	"github.com/allegro/marathon-appcop/marathon"
	"github.com/allegro/marathon-appcop/metrics"
	"github.com/allegro/marathon-appcop/mgc"
	"github.com/allegro/marathon-appcop/score"
	"github.com/allegro/marathon-appcop/web"
)

// Version variable provided at build time
var Version string

func main() {

	log.Infof("Appcop Version: %s", Version)
	config, err := config.NewConfig()
	if err != nil {
		log.Fatal(err.Error())
	}

	err = metrics.Init(config.Metrics)
	if err != nil {
		log.Fatal(err.Error())
	}

	remote, err := marathon.New(config.Marathon)
	if err != nil {
		log.Fatal(err.Error())
	}

	scores, err := score.New(config.Score, remote)
	if err != nil {
		log.Fatal(err.Error())
	}
	updates := scores.ScoreManager()

	gc, err := mgc.New(config.MGC, remote)
	if err != nil {
		log.Fatal(err.Error())
	}
	stop := web.NewHandler(config.Web, remote, gc, updates)
	defer stop()

	// set up routes
	http.HandleFunc("/health", web.HealthHandler)

	log.WithField("Port", config.Web.Listen).Info("Listening")
	log.Fatal(http.ListenAndServe(config.Web.Listen, nil))

}
