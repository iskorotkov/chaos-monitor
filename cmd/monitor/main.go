package main

import (
	"context"
	"log"
	"os"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	_ "go.uber.org/automaxprocs"
	v1 "k8s.io/api/core/v1"

	"github.com/iskorotkov/chaos-monitor/pkg/analyzer"
	"github.com/iskorotkov/chaos-monitor/pkg/kube"
	"github.com/iskorotkov/chaos-monitor/pkg/parser"
)

var (
	appNS       = os.Getenv("APP_NS")
	runDuration = os.Getenv("DURATION")

	ignoredPods   = parser.AsSet(os.Getenv("IGNORED_PODS"), ";")
	ignoredLabels = parser.AsSet(os.Getenv("IGNORED_LABELS"), ";")
	ignoredNodes  = parser.AsSet(os.Getenv("IGNORED_NODES"), ";")

	redisAddr    = os.Getenv("REDIS_ADDR")
	redisChannel = os.Getenv("REDIS_CHANNEL")
	errorsLimit  = mustParseInt(os.Getenv("ERRORS_LIMIT"))
)

func main() {
	// Handle panics.
	defer func() {
		r := recover()
		if r != nil {
			log.Printf("panic occurred: %v", r)
			debug.PrintStack()
			os.Exit(1)
		}
	}()

	redisClient := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer redisClient.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	ping := redisClient.Ping(ctx)
	if ping.Err() != nil {
		log.Printf("failed to connect to redis: %v", ping.Err())
		return
	}

	go monitorRedisForErrors(redisClient)

	if appNS == "" {
		appNS = "default"
	}

	logger := log.New(log.Writer(), log.Prefix(), log.Flags())
	failureDetector := analyzer.NewAnalyzer(ignoredPods, ignoredLabels, ignoredNodes, logger)
	kube.StartMonitor(appNS, runDuration, lookForFailures(failureDetector))
}

// lookForFailures outputs pod event messages.
func lookForFailures(counter analyzer.Analyzer) kube.OnUpdateFunction {
	return func(_, newObj interface{}) {
		pod, ok := newObj.(*v1.Pod)
		if !ok {
			log.Fatal("couldn't cast object to pod")
		}

		err := counter.Analyze((*analyzer.Pod)(pod))
		if err != nil {
			log.Fatal(err)
		}
	}
}

func monitorRedisForErrors(redisClient *redis.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	sub := redisClient.Subscribe(ctx, redisChannel)
	defer sub.Close()

	var errors int
	for msg := range sub.Channel() {
		log.Printf("received message: %s", msg.Payload)

		errors++
		if errors > errorsLimit {
			log.Printf("too many errors, exiting")
			os.Exit(1)
		}
	}
}

func mustParseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Fatal(err)
	}

	return i
}
