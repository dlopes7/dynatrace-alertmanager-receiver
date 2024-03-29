package jobs

import (
	"fmt"
	"github.com/dlopes7/dynatrace-alertmanager-receiver/pkg/cache"
	dtapi "github.com/dlopes7/dynatrace-go-client/api"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

type Scheduler struct {
	customDeviceCache *cache.CustomDeviceCacheService
	problemCache      *cache.ProblemCacheService
	dtClient          dtapi.Client
}

func NewScheduler(deviceCache *cache.CustomDeviceCacheService, problemCache *cache.ProblemCacheService) Scheduler {
	dt := dtapi.New(dtapi.Config{
		APIKey:    os.Getenv("DT_API_TOKEN"),
		BaseURL:   os.Getenv("DT_API_URL"),
		Retries:   5,
		RetryTime: 2 * time.Second,
	})
	return Scheduler{
		dtClient:          dt,
		customDeviceCache: deviceCache,
		problemCache:      problemCache,
	}
}

// UpdateProblemIDs checks for alerts without a ProblemID in the cache, and update them with their ProblemIDs
func (s *Scheduler) UpdateProblemIDs() {
	log.Info("Scheduler - Starting UpdateProblemIDs")
	s.problemCache.Lock()
	problemCache := s.problemCache.GetCache()

	// Copy the map so that we can update this during the iteration below
	updatedProblems := map[string]cache.Problem{}
	for hash, problem := range problemCache.Problems {
		updatedProblems[hash] = problem
	}

	dtProblems, _, err := s.dtClient.Problem.ListV1("", 0, 0, "OPEN", "", "", nil, true)
	if err != nil {
		log.WithFields(log.Fields{"error": err.Error()}).Error("Scheduler - Error obtaining Dynatrace Problems")
	} else {
		for hash, problem := range problemCache.Problems {
			foundProblem := false

			if problem.ProblemID == "" {
				// Only look for the ProblemID if we don't have it already

				entity := problem.Event.AttachRules.EntityIds[0]
				correlationID := problem.EventStoreResult.StoredCorrelationIds[0]
				log.WithFields(log.Fields{"hash": hash, "entity": entity, "alert": problem.Event.Title, "correlationID": correlationID}).Info("Scheduler - Found an alert without a ProblemID")

				// Problems V1 API gives us the correlationID for each event, we just compare the values for each event of opened problem to find ours
				for _, dtProblem := range dtProblems {
					for _, event := range dtProblem.RankedEvents {
						if event.CorrelationID == correlationID {
							log.WithFields(log.Fields{"hash": hash, "entity": entity, "problem": dtProblem.ID, "correlationID": correlationID}).Info("Scheduler - Found a ProblemID for the event")
							problem.ProblemID = dtProblem.ID
							updatedProblems[hash] = problem
							foundProblem = true
						}
					}
				}
				if foundProblem == false {
					log.WithFields(log.Fields{"hash": hash}).Warning("Scheduler - Could not find a Problem with an event matching the hash")
				}
			}
		}
	}

	problemCache.Problems = updatedProblems
	s.problemCache.Update(*problemCache)
	s.problemCache.UnLock()

}

func (s *Scheduler) ResendEvents() {

	log.Info("Scheduler - Starting ResendEvents")
	s.problemCache.Lock()
	problemCache := s.problemCache.GetCache()
	for _, problem := range problemCache.Problems {
		r, _, err := s.dtClient.Events.Create(problem.Event)
		if err != nil {
			log.WithFields(log.Fields{"error": err.Error()}).Error("Scheduler - Could not resent the event")
		}
		log.WithFields(log.Fields{"response": fmt.Sprintf("%+v", r)}).Info("Scheduler - Dynatrace response after sending the event")
	}
	s.problemCache.UnLock()

}

func (s *Scheduler) DeleteOldEvents() {

	now := time.Now()
	log.Info("Scheduler - Starting DeleteOldEvents")
	s.problemCache.Lock()
	problemCache := s.problemCache.GetCache()
	for hash, problem := range problemCache.Problems {
		timeAlive := now.Sub(problem.CreatedAt)
		if timeAlive > 5*24*time.Hour {
			log.WithFields(log.Fields{"CreatedAt": problem.CreatedAt, "timeAlive": timeAlive}).Info("Scheduler - Deleting event because it is too old")
			s.problemCache.Delete(hash)
		}
	}
	s.problemCache.UnLock()
}
