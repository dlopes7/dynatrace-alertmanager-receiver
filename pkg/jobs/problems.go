package jobs

import (
	"github.com/dlopes7/dynatrace-alertmanager-receiver/pkg/cache"
	dtapi "github.com/dyladan/dynatrace-go-client/api"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
	"time"
)

type ProblemJob struct {
	customDeviceCache *cache.CustomDeviceCacheService
	problemCache      *cache.ProblemCacheService
	dtClient          dtapi.Client
}

func NewProblemJob(deviceCache *cache.CustomDeviceCacheService, problemCache *cache.ProblemCacheService) ProblemJob {
	dt := dtapi.New(dtapi.Config{
		APIKey:    os.Getenv("DT_API_KEY"),
		BaseURL:   os.Getenv("DT_BASE_URL"),
		Retries:   5,
		RetryTime: 2 * time.Second,
	})
	return ProblemJob{
		dtClient:          dt,
		customDeviceCache: deviceCache,
		problemCache:      problemCache,
	}
}

// UpdateProblemIDs checks for alerts without a ProblemID in the cache, and update them with their ProblemIDs
func (p *ProblemJob) UpdateProblemIDs() {

	fields := []string{"evidenceDetails"}
	problemSelector := "status(\"open\")"

	problemCache := p.problemCache.GetCache()

	// Copy the map so that we can update this during the iteration below
	updatedProblems := map[string]cache.Problem{}
	for hash, problem := range problemCache.Problems {
		updatedProblems[hash] = problem
	}

	for hash, problem := range problemCache.Problems {
		if problem.ProblemID == "" {
			entity := problem.Event.AttachRules.EntityIds[0]
			log.WithFields(log.Fields{"hash": hash, "entity": entity, "alert": problem.Event.Title}).Info("ProblemJob - Found an alert without a ProblemID")

			dtProblems, _, err := p.dtClient.Problem.List(fields, problemSelector, "", "")
			if err != nil {
				log.WithFields(log.Fields{"error": err.Error()}).Info("ProblemJob - Error obtaining Dynatrace Problems")
			} else {
				foundProblem := false
				for _, dtProblem := range dtProblems {
					for _, evidenceDetails := range dtProblem.EvidenceDetails.Details {
						if strings.Contains(evidenceDetails.DisplayName, hash) {
							log.WithFields(log.Fields{"hash": hash, "entity": entity, "problem": dtProblem.ProblemID}).Info("ProblemJob - Found a ProblemID for the event")
							problem.ProblemID = dtProblem.ProblemID
							updatedProblems[hash] = problem
							foundProblem = true
						}
					}
				}
				if foundProblem == false {
					log.WithFields(log.Fields{"hash": hash}).Warning("ProblemJob - Could not find a Problem with an event matching the hash")
				}
			}
		}
	}
	problemCache.Problems = updatedProblems
	p.problemCache.Update(*problemCache)

}
