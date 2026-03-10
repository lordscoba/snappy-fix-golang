package redis

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/snappy-fix-golang/internal/adapters"
	"github.com/snappy-fix-golang/internal/adapters/db"
	"github.com/snappy-fix-golang/internal/config"

	logutil "github.com/snappy-fix-golang/pkg/logger"
)

var (
	Ctx     = context.Background()
	KeyName = "EmailQueue"
)

func ConnectToRedis(logger *logutil.Logger, configDatabases config.Redis) *redis.Client {
	dbsCV := configDatabases
	logutil.LogAndPrint(logger, "connecting to redis server")
	connectedServer := connectToDb(dbsCV.REDIS_HOST, dbsCV.REDIS_PORT, dbsCV.REDIS_DB, logger)
	logutil.LogAndPrint(logger, "connected to redis server")
	db.DB.Redis = NewRedisConnection(connectedServer)
	return connectedServer
}

func connectToDb(host, port, db string, logger *logutil.Logger) *redis.Client {
	port = adapters.ResolvePortParsing(port, logger)
	dbInst, err := strconv.Atoi(db)
	if err != nil {
		logutil.LogAndPrint(logger, fmt.Sprintf("parsing url %v to get port failed with: %v", port, err))
		panic(err)
	}

	addr := fmt.Sprintf("%v:%v", host, port)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       dbInst,
	})

	if err := redisClient.Ping(Ctx).Err(); err != nil {
		logutil.LogAndPrint(logger, fmt.Sprintln(addr))
		logutil.LogAndPrint(logger, fmt.Sprintln("Redis db error: ", err))
	}

	pong, _ := redisClient.Ping(Ctx).Result()
	logutil.LogAndPrint(logger, fmt.Sprintln("Redis says: ", pong))

	return redisClient
}
