package main

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

type Service struct {
	db *redis.Client
}

type Dependency struct {
	name    string
	version string
	url     string
}

func (dep *Dependency) Key() string {
	return dep.name + "@" + dep.version
}

const VASGO_URL = "localhost:6379"

func NewService() *Service {
	client := redis.NewClient(&redis.Options{
		Addr:     VASGO_URL,
		Password: "",
		DB:       0,
	})

	return &Service{db: client}
}

func (s *Service) GetAliveDependencies(deps map[string]Dependency, aliveDeps map[string]Dependency) map[string]Dependency {
	if len(deps) == 0 {
		return aliveDeps
	}

	for name, dep := range deps {
		key := dep.Key()
		url, err := s.db.SRandMember(key).Result()

		if err != nil {
			panic("Can not find dependency for " + key)
		}

		alive, err := s.db.Get("alives." + url).Result()

		if err != nil {
			panic("Can not find alive url for " + key)
		}

		if alive == dep.Key() {
			aliveDeps[dep.name] = Dependency{
				name:    name,
				version: dep.version,
				url:     url,
			}
			delete(deps, dep.name)
		} else {
			_, err := s.db.SRem(dep.Key(), url).Result()
			if err != nil {
				panic("can not remove not alive url")
			}
		}
	}

	return s.GetAliveDependencies(deps, aliveDeps)
}

func (s *Service) FindDependencies(dependencies []Dependency) ([]Dependency, error) {
	deps := make(map[string]Dependency)

	for _, dep := range dependencies {
		deps[dep.name] = dep
	}

	aliveDeps := s.GetAliveDependencies(deps, make(map[string]Dependency))

	i := 0
	result := make([]Dependency, len(aliveDeps))

	for _, dep := range aliveDeps {
		result[i] = dep
		i++
	}

	return result, nil
}

func (s *Service) Register(service Dependency) (*Service, error) {
	pkey := "endpoints." + service.Key()
	akey := "alives." + service.url

	s.db.SAdd(pkey, service.url)

	duration, _ := time.ParseDuration("10s")
	s.db.Set(akey, service.Key(), duration)

	return s, nil
}

func main() {
	service := NewService()

	app := Dependency{
		name:    "app",
		version: "0.0.1",
		url:     "app.beansauce.io",
	}

	app2 := Dependency{
		name:    "app",
		version: "0.0.1",
		url:     "beta.app.beansauce.io",
	}

	web := Dependency{
		name:    "web",
		version: "0.1.2",
		url:     "web.beansauce.io",
	}

	db := Dependency{
		name:    "db",
		version: "0.1.3",
		url:     "db.beansauce.io",
	}

	service.Register(app)
	service.Register(app2)
	service.Register(web)
	service.Register(db)

	dependencies := []Dependency{app, web, db}

	result, _ := service.FindDependencies(dependencies)

	for _, r := range result {
		fmt.Printf("result: %v@%v => %v\n", r.name, r.version, r.url)
	}
}
