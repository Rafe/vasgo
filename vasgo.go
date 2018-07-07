package vasgo

import (
	"time"

	"github.com/go-redis/redis"
)

type Service struct {
	db *redis.Client
}

type Endpoint struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Url     string `json:"url"`
	Alive   bool   `json:"alive"`
}

func NewEndpoint(name string, version string, url string, alive bool) *Endpoint {
	return &Endpoint{
		Name:    name,
		Version: version,
		Url:     url,
		Alive:   alive,
	}
}

func (p *Endpoint) Key() string {
	return p.Name + "@" + p.Version
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

	if All(endpoints, func(p Endpoint) bool { return p.Alive == true }) {
		return endpoints, nil
	}

	result := make([]Endpoint, len(endpoints))

	for i, p := range endpoints {
		url, err := s.db.SRandMember(p.Key()).Result()

		if err != nil || url == "" {
			panic("Can not find dependency for " + p.Key())
		}

		result[i] = Endpoint{
			Name:    p.Name,
			Version: p.Version,
			Url:     url,
		}
	}

	for i, p := range result {
		alive, err := s.db.Get("alives." + p.Url).Result()

		if err != nil {
			panic("Can not find alive url for " + p.Key())
		}

		if alive == p.Key() {
			result[i].Alive = true
		} else {
			result[i].Alive = false

			_, err := s.db.SRem(p.Key(), p.Url).Result()
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
			Name:    key,
			Version: val,
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
	s.db.SAdd(pkey, p.Url)

	go s.SetEndpointHealth(p)

	return s, nil
}

func (s *Service) SetEndpointHealth(p *Endpoint) error {
	akey := "alives." + p.Url

	aliveDuration, _ := time.ParseDuration("10s")
	s.db.Set(akey, p.Key(), aliveDuration)

	select {
	case <-time.After(aliveDuration):
		go s.SetEndpointHealth(p)
	}

	return nil
}
