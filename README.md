# Go Vasgo

## What is Vasgo?

Vasgo is a practice/experiment project for service discovery library using Redis, concept is port from [node-vasco](https://github.com/asyncanup/vasco)

## Problems

+ Service discovery for micro-services is too hard to understand
+ Most solution depends on software like Zookeeper or DNS
+ No good standalone example in Go

## Install

go get -u https://github.com/Rafe/vasgo

## Usage


Register
=======
```go
package main

import (
    "github.com/rafe/vasgo"
)

func main() {
    const REDIS_URL = "localhost:6379"
    service := vasgo.NewService(REDIS_URL, "")

    app := vasgo.NewEndpoint("app", "0.0.1", "https://168.0.0.1")

    _, err := vasgo.Register(app)

    if err != nil {
        panic("Failed to register service")
    }
}
```

Find dependencies
=======
```go
package main

import (
    "fmt"

    "github.com/rafe/vasgo"
)

func main() {
    const REDIS_URL = "localhost:6379"
    service := vasgo.NewService(REDIS_URL, "")

    dependencies := map[string]string{
      "app": "0.0.1",
      "web": "0.1.2",
      "db":  "0.1.3",
    }

    endpoints, err := vasgo.FindDependencies(dependencies)

    if err != nil {
        panic("Failed to find dependencies")
    }

    for _, p := range endpoints {
        fmt.Println(p)
    }
    // {app 0.0.1 url... alive}
    // {web 0.1.2 url... alive}
    // {db  0.1.3 url... alive}
}
```
