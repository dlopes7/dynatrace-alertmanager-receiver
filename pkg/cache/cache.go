package cache

import (
	"encoding/json"
	"fmt"
	"github.com/dlopes7/dynatrace-alertmanager-receiver/alertmanager"
	"github.com/dlopes7/dynatrace-alertmanager-receiver/pkg/utils"
	dynatrace "github.com/dyladan/dynatrace-go-client/api"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"sync"
	"time"
)

type CustomDeviceCacheService struct {
	cache    CustomDeviceCache
	lock     sync.Mutex
	location string
}

type CustomDeviceCache struct {
	CustomDevices []string  `json:"customDevices"`
	LastUpdated   time.Time `json:"lastUpdated"`
}

func NewCustomDeviceCacheService() CustomDeviceCacheService {
	cache := CustomDeviceCache{
		CustomDevices: []string{},
		LastUpdated:   time.Now(),
	}
	return CustomDeviceCacheService{
		location: fmt.Sprintf("%s/customDevices.json", utils.GetTempDir()),
		cache:    cache,
	}
}

func (c *CustomDeviceCacheService) GetCache() *CustomDeviceCache {
	var cache CustomDeviceCache
	c.lock.Lock()
	jsonFile, err := os.Open(c.location)
	if err != nil {
		log.WithFields(log.Fields{"location": c.location, "error": err.Error()}).Warning("Could not open custom device cache file, will create a new one")
		cache = c.cache
	} else {
		defer jsonFile.Close()
		byteValue, _ := ioutil.ReadAll(jsonFile)
		err = json.Unmarshal(byteValue, &cache)
		if err != nil {
			log.WithFields(log.Fields{"location": c.location, "error": err.Error()}).Warning("Could not parse the custom device cache file, resetting the cache")
			cache = c.cache
		}
	}
	c.lock.Unlock()
	return &cache
}

func (c *CustomDeviceCacheService) Update(cd CustomDeviceCache) {
	c.lock.Lock()
	c.cache = cd
	c.cache.LastUpdated = time.Now()
	file, _ := json.MarshalIndent(c.cache, "", " ")
	_ = ioutil.WriteFile(c.location, file, 0644)
	c.lock.Unlock()
}

type ProblemCacheService struct {
	cache    ProblemCache
	lock     sync.Mutex
	location string
}

type ProblemCache struct {
	Problems    map[string]Problem `json:"problems"`
	LastUpdated time.Time          `json:"lastUpdated"`
}

type Problem struct {
	Event            dynatrace.EventCreation    `json:"event"`
	Alert            alertmanager.Data          `json:"alert"`
	CreatedAt        time.Time                  `json:"createdAt"`
	EventStoreResult dynatrace.EventStoreResult `json:"eventStoreResult"`
	ProblemID        string                     `json:"problemID"`
}

func NewProblemCacheService() ProblemCacheService {
	pc := ProblemCache{
		Problems:    map[string]Problem{},
		LastUpdated: time.Now(),
	}
	return ProblemCacheService{
		location: fmt.Sprintf("%s/problems.json", utils.GetTempDir()),
		cache:    pc,
	}
}

func (p *ProblemCacheService) GetCache() *ProblemCache {
	var cache ProblemCache
	p.lock.Lock()
	jsonFile, err := os.Open(p.location)
	if err != nil {
		log.WithFields(log.Fields{"location": p.location, "error": err.Error()}).Warning("Could not open problems cache file, will create a new one")
		cache = p.cache
	} else {
		defer jsonFile.Close()
		byteValue, _ := ioutil.ReadAll(jsonFile)
		err = json.Unmarshal(byteValue, &cache)
		if err != nil {
			log.WithFields(log.Fields{"location": p.location, "error": err.Error()}).Warning("Could not parse the problem cache file, resetting the cache")
			cache = p.cache
		}
	}
	p.lock.Unlock()
	return &cache
}

func (p *ProblemCacheService) AddProblem(hash string, problem Problem) {
	cache := p.GetCache()
	p.lock.Lock()
	cache.Problems[hash] = problem
	cache.LastUpdated = time.Now()
	p.cache = *cache
	p.persist()
	p.lock.Unlock()
}

func (p *ProblemCacheService) Update(pc ProblemCache) {
	p.lock.Lock()
	for hash, problem := range pc.Problems {
		p.cache.Problems[hash] = problem
	}
	p.cache.LastUpdated = time.Now()
	p.persist()
	p.lock.Unlock()
}

func (p *ProblemCacheService) persist() {
	file, _ := json.MarshalIndent(p.cache, "", " ")
	_ = ioutil.WriteFile(p.location, file, 0644)

}

func (p *ProblemCacheService) Delete(hash string) {
	log.WithFields(log.Fields{"hash": hash}).Info("ProblemCacheService - deleting the cache entry")
	p.lock.Lock()
	delete(p.cache.Problems, hash)
	p.cache.LastUpdated = time.Now()
	p.persist()
	p.lock.Unlock()
}
