package redis

import (
	"github.com/go-redis/redis"
	"strconv"
	"time"
)

// _RedisPool 类型， 是一个database容器，用于存储服务可能用到的所有redis连接池
type _RedisPool struct {
	pool map[string]*rdbQuery
}

// MysqlQuery 对象，用于直接查询或执行
type rdbQuery struct {
	client *redis.Client
	alias  string
}

type rdbConnector interface {
	Set(key string, val interface{})
}

var redisPool *_RedisPool

func init() {
	if redisPool != nil {
		return
	}
	log.Info("init redis module...")
	redisPool = &_RedisPool{}
	redisPool.pool = make(map[string]*rdbQuery)
}

func AddRedis(name, addr string, DB int) {
	log.Info("add redis", name, addr+"/"+strconv.Itoa(DB))
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       DB,
	})
	if _, err := client.Ping().Result(); err != nil {
		log.Error("can not ping redis", addr, "DB:"+strconv.Itoa(DB), err.Error())
		panic(err.Error()) // 直接panic，让server无法启动
	}
	rdQuery := &rdbQuery{}
	rdQuery.client = client
	rdQuery.alias = name
	redisPool.pool[name] = rdQuery
}

func UseRedis(name string) *rdbQuery {
	return redisPool.pool[name]
}

func (rdq *rdbQuery) Get(key string) (string, error) {
	log.Debug("[get redis]", "[redis: "+rdq.alias+"]", "GET ", key)
	return rdq.client.Get(key).Result()
}

func (rdq *rdbQuery) Set(key, value string, expiration time.Duration) error {
	log.Debug("[set redis]", "[redis: "+rdq.alias+"]", "SET ", key, value)
	return rdq.client.Set(key, value, expiration).Err()
}

func (rdq *rdbQuery) SetNX(key, value string, expiration time.Duration) error {
	// 当且仅当key不存在时，将key的值设为value
	log.Debug("[setNX redis]", "[redis: "+rdq.alias+"]", "SETNX ", key, value)
	return rdq.client.SetNX(key, value, expiration).Err()
}

func (rdq *rdbQuery) SetXX(key, value string, expiration time.Duration) error {
	// 当且仅当key存在时，将key的值设为value
	log.Debug("[setXX redis]", "[redis: "+rdq.alias+"]", "SETXX ", key, value)
	return rdq.client.SetXX(key, value, expiration).Err()
}

func (rdq *rdbQuery) Incr(key string) (int64, error) {
	log.Debug("[incr redis]", "[redis: "+rdq.alias+"]", "Incr ", key)
	return rdq.client.Incr(key).Result()
}

func (rdq *rdbQuery) Decr(key string) (int64, error) {
	log.Debug("[decr redis]", "[redis: "+rdq.alias+"]", "Decr ", key)
	return rdq.client.Decr(key).Result()
}

func (rdq *rdbQuery) Del(key string) error {
	log.Debug("[del redis]", "[redis: "+rdq.alias+"]", "Del ", key)
	return rdq.client.Del(key).Err()
}

func (rdq *rdbQuery) ExpireAt(key string, tm time.Time) error {
	log.Debug("[expireat redis]", "[redis: "+rdq.alias+"]", "ExpireAt", key)
	return rdq.client.ExpireAt(key, tm).Err()
}

func (rdq *rdbQuery) Expire(key string, expiration time.Duration) error {
	log.Debug("[expire redis]", "[redis: "+rdq.alias+"]", "Expire", key)
	return rdq.client.Expire(key, expiration).Err()
}

func (rdq *rdbQuery) Exists(key ...string) int64 {
	log.Debug("[exists redis]", "[redis: "+rdq.alias+"]", "Exists", key)
	return rdq.client.Exists(key).Val()
}
