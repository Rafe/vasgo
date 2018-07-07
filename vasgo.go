package vasgo

import (
	"time"

	"github.com/go-redis/redis"
)

type Service struct {
	db *redis.Client
}

type Endpoint struct {
	name    string
	version string
	url     string
	alive   bool
}

func NewEndpoint(name string, version string, url string, alive bool) *Endpoint {
	return &Endpoint{
		name:    name,
		version: version,
		url:     url,
		alive:   alive,
	}
}

func (p *Endpoint) Key() string {
	return p.name + "@" + p.version
}

func NewService(url string, password string) *Service {
	client := redis.NewClient(&redis.Options{
		Addr:     url,
		Password: password,
		DB:       0,
	})

	return &Service{db: client}
}

func (s *Service) GetAliveEndpoints(endpoints []Endpoint) ([]Endpoint, error) {
	if len(endpoints) == 0 {
		return endpoints, nil
	}

	if All(endpoints, func(p Endpoint) bool { return p.alive == true }) {
		return endpoints, nil
	}

	result := make([]Endpoint, len(endpoints))

	for i, p := range endpoints {
		url, err := s.db.SRandMember(p.Key()).Result()

		if err != nil || url == "" {
			panic("Can not find dependency for " + p.Key())
		}

		result[i] = Endpoint{
			name:    p.name,
			version: p.version,
			url:     url,
		}
	}

	for i, p := range result {
		alive, err := s.db.Get("alives." + p.url).Result()

		if err != nil {
			panic("Can not find alive url for " + p.Key())
		}

		if alive == p.Key() {
			result[i].alive = true
		} else {
			result[i].alive = false

			_, err := s.db.SRem(p.Key(), p.url).Result()
			if err != nil {
				panic("can not remove not alive url")
			}
		}
	}

	return s.GetAliveEndpoints(result)
}

func All(collection []Endpoint, f func(Endpoint) bool) bool {
	for _, v := range collection {
		if !f(v) {
			return false
		}
	}

	return true
}

func (s *Service) FindDependencies(deps map[string]string) ([]Endpoint, error) {
	endpoints := make([]Endpoint, len(deps))

	if len(endpoints) == 0 {
		return endpoints, nil
	}

	i := 0
	for key, val := range deps {
		endpoints[i] = Endpoint{
			name:    key,
			version: val,
		}
		i++
	}

	result, err := s.GetAliveEndpoints(endpoints)

	if err != nil {
		panic("Error on finding endpoints")
	}

	return result, nil
}

func (s *Service) Register(p *Endpoint) (*Service, error) {
	pkey := "endpoints." + p.Key()
	akey := "alives." + p.url

	s.db.SAdd(pkey, p.url)

	duration, _ := time.ParseDuration("10s")
	s.db.Set(akey, p.Key(), duration)

	return s, nil
}
