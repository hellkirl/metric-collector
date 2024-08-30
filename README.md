# MemStats and alerting collection service

<img src="https://www.agilitypr.com/wp-content/uploads/2023/01/metric.jpg">

## Content

1. [Launch Manual](#launch-manual)
2. [Database Description](#database-description)
3. [Project Structure Description](#project-structure-description)

<img src="https://raw.githubusercontent.com/andreasbm/readme/master/assets/lines/rainbow.png">

## Launch manual

<img src="https://raw.githubusercontent.com/andreasbm/readme/master/assets/lines/rainbow.png">

## Database Description

<img src="https://raw.githubusercontent.com/andreasbm/readme/master/assets/lines/rainbow.png">

## Project Structure Description

```
.
├── README.md
├── cmd
│   ├── agent
│   │   ├── main.go
│   │   └── main_test.go
│   └── server
│       └── main.go
├── go.mod
├── go.sum
├── internal
│   ├── handlers
│   │   └── server
│   │       ├── handlers.go
│   │       └── handlers_test.go
│   ├── storage
│   │   └── metricsStorage.go
│   └── util
└── static
    └── metric.html
```