package db

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"k8s.io/klog/v2"
)

var RedisClient *redis.Client

var lock sync.RWMutex

type RedisConfig struct {
	Address     string        `json:"address"`
	Password    string        `json:"pwd"`
	PoolSize    int           `json:"poolsize"`
	MaxIdle     int           `json:"maxidle"`
	IdleTimeOut time.Duration `json:"idletimeout"`
}

func (r *RedisConfig) Init() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:            r.Address,
		Password:        r.Password,
		PoolSize:        r.PoolSize,
		MaxIdleConns:    r.MaxIdle,
		ConnMaxIdleTime: r.IdleTimeOut,
	})

	ctx := context.Background()
	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		klog.Error("redis connect failed:", err)
		panic(err)
	}
}

// duration = 5 * time.Minute
// expire := duration.String()
// command[]= [command, key, value, expire]
func Dao(command ...string) (string, error) {
	ctx := context.Background()
	cn := RedisClient.Conn()
	defer cn.Close()

	switch command[0] {
	case "get":
		key := command[1]
		lock.RLock()
		val, err := cn.Get(ctx, key).Result()
		lock.RUnlock()
		if err != nil {
			return "", errors.New("Redis command:  Get " + key + " failed!")
		}
		return val, nil
	// set: if key exist, will update value
	case "set":
		key, value := command[1], command[2]
		expire, err := time.ParseDuration(command[3])
		if err != nil {
			klog.Error("redis expire-time parse failed!")
			return "", errors.New("Redis command:  Set " + key + " failed!")
		}

		lock.Lock()
		err = cn.Set(ctx, key, value, expire).Err()
		lock.Unlock()
		if err != nil {
			return "", errors.New("Redis command:  Set " + key + " failed!")
		}
		return "", nil
	case "del":
		key := command[1]
		lock.Lock()
		err := cn.Del(ctx, key).Err()
		lock.Unlock()
		if err != nil {
			return "", errors.New("Redis command:  Del " + key + " failed!")
		}
		return "", nil
	default:
		return "", errors.New("Command: " + command[0] + " not support!")
	}
}
