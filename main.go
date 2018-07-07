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
	alive   bool
}

func (dep *Dependency) Key() string {
	return dep.name + "@" + dep.version
}

const VASGO_URL = "localhost:6379"

func NewService(url string, password string) *Service {
	client := redis.NewClient(&redis.Options{
		Addr:     url,
		Password: password,
		DB:       0,
	})

	return &Service{db: client}
}

func (s *Service) GetAliveDependencies(deps []Dependency) ([]Dependency, error) {
	if len(deps) == 0 {
		return deps, nil
	}

	if All(deps, func(dep Dependency) bool { return dep.alive == true }) {
		return deps, nil
	}

	result := make([]Dependency, len(deps))

	for i, dep := range deps {
		url, err := s.db.SRandMember(dep.Key()).Result()

		if err != nil || url == "" {
			panic("Can not find dependency for " + dep.Key())
		}

		result[i] = Dependency{
			name:    dep.name,
			version: dep.version,
			url:     url,
		}
	}

	for i, dep := range deps {
		alive, err := s.db.Get("alives." + dep.url).Result()

		if err != nil {
			panic("Can not find alive url for " + dep.Key())
		}

		if alive == dep.Key() {
			result[i].alive = true
		} else {
			result[i].alive = false

			_, err := s.db.SRem(dep.Key(), dep.url).Result()
			if err != nil {
				panic("can not remove not alive url")
			}
		}
	}

	return s.GetAliveDependencies(result)
}

func All(collection []Dependency, f func(Dependency) bool) bool {
	for _, v := range collection {
		if !f(v) {
			return false
		}
	}

	return true
}

func (s *Service) FindDependencies(deps []Dependency) ([]Dependency, error) {
	if len(deps) == 0 {
		return deps, nil
	}

	result, err := s.GetAliveDependencies(deps)

	if err != nil {
		panic("Error on finding alive dependencies")
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
	service := NewService(VASGO_URL, "")

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
		fmt.Printf("result: %v@%v => %v, alive? %v\n", r.name, r.version, r.url, r.alive)
	}

}
