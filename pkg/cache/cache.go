package cache

import (
	"encoding/json"
	"fmt"
	"github.com/dlopes7/dynatrace-alertmanager-receiver/alertmanager"
	"github.com/dlopes7/dynatrace-alertmanager-receiver/pkg/utils"
	dynatrace "github.com/dlopes7/dynatrace-go-client/api"
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

type CustomDevice struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Group string `json:"group"`
}

type CustomDeviceCache struct {
	CustomDevices []CustomDevice `json:"customDevices"`
	LastUpdated   time.Time      `json:"lastUpdated"`
}

type CustomDeviceCacheV1 struct {
	CustomDevices []string  `json:"customDevices"`
	LastUpdated   time.Time `json:"lastUpdated"`
}

func (c *CustomDeviceCache) GetIDs() []string {
	var ids []string
	for _, cd := range c.CustomDevices {
		ids = append(ids, cd.ID)
	}
	return ids

}

func NewCustomDeviceCacheService() CustomDeviceCacheService {
	cache := CustomDeviceCache{
		CustomDevices: []CustomDevice{},
		LastUpdated:   time.Now(),
	}
	return CustomDeviceCacheService{
		location: fmt.Sprintf("%s/customDevices.json", utils.GetTempDir()),
		cache:    cache,
	}
}

func (c *CustomDeviceCacheService) updateCacheFromV1(dtClient *dynatrace.Client) (*CustomDeviceCache, error) {
	// Necessary because I've changed the format of the cache
	// If we find a cache on the old format, convert it to the new one
	var cache CustomDeviceCacheV1
	jsonFile, err := os.Open(c.location)
	if err != nil {
		log.WithFields(log.Fields{"location": c.location, "error": err.Error()}).Warning("Could not update the cache")
		return nil, err
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	err = json.Unmarshal(byteValue, &cache)
	if err != nil {
		log.WithFields(log.Fields{"location": c.location, "error": err.Error()}).Warning("Could not update the custom device cache file")
		return nil, err
	}
	// create a CustomDeviceCache with the devices from the current cache
	var customDevices []CustomDevice
	for _, id := range cache.CustomDevices {
		name := id
		log.WithFields(log.Fields{"id": id}).Info("Attempting to update the custom device name")

		entity, _, err := dtClient.Entities.Get(id)
		if err == nil {
			name = entity.DisplayName
			log.WithFields(log.Fields{"id": id, "name": name}).Info("Setting the custom device name")
		}
		customDevices = append(customDevices, CustomDevice{ID: id, Name: name, Group: os.Getenv("DT_GROUP_NAME")})
	}
	return &CustomDeviceCache{
		CustomDevices: customDevices,
	}, nil

}

func (c *CustomDeviceCacheService) GetCache(dtClient *dynatrace.Client) *CustomDeviceCache {
	var cache CustomDeviceCache
	jsonFile, err := os.Open(c.location)
	if err != nil {
		log.WithFields(log.Fields{"location": c.location, "error": err.Error()}).Warning("Could not open custom device cache file, will create a new one")
		cache = c.cache
	} else {
		defer jsonFile.Close()
		byteValue, _ := ioutil.ReadAll(jsonFile)
		err = json.Unmarshal(byteValue, &cache)
		if err != nil {
			log.WithFields(log.Fields{"location": c.location, "error": err.Error()}).Warning("Could not parse the custom device cache file, attempting to update")
			updatedCache, err := c.updateCacheFromV1(dtClient)
			if err != nil {
				log.WithFields(log.Fields{"location": c.location, "error": err.Error()}).Warning("Could not update the custom device cache file, resetting the cache")
				cache = c.cache
			} else {
				cache = *updatedCache
				c.Update(cache)
			}
		}
	}
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
	return &cache
}

func (p *ProblemCacheService) Lock() {
	p.lock.Lock()
}

func (p *ProblemCacheService) UnLock() {
	p.lock.Unlock()
}

func (p *ProblemCacheService) AddProblem(hash string, problem Problem) {
	p.lock.Lock()
	cache := p.GetCache()
	cache.Problems[hash] = problem
	cache.LastUpdated = time.Now()
	p.cache = *cache
	p.persist()
	p.lock.Unlock()
}

func (p *ProblemCacheService) Update(pc ProblemCache) {
	for hash, problem := range pc.Problems {
		p.cache.Problems[hash] = problem
	}
	p.cache.LastUpdated = time.Now()
	p.persist()
}

func (p *ProblemCacheService) persist() {
	file, _ := json.MarshalIndent(p.cache, "", " ")
	_ = ioutil.WriteFile(p.location, file, 0644)

}

func (p *ProblemCacheService) Delete(hash string) {
	log.WithFields(log.Fields{"hash": hash}).Info("ProblemCacheService - deleting the cache entry")
	delete(p.cache.Problems, hash)
	p.cache.LastUpdated = time.Now()
	p.persist()
}
