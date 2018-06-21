package redis

import (
	"testing"
	"time"
)

func init() {
	AddRedis("test", "127.0.0.1:6379", 0)
}

func TestRedis(t *testing.T) {
	client := UseRedis("test")
	client.Set("key1", "value1", 60*time.Second)
	client.SetNX("key2", "value2", 60*time.Second)

	if val, err := client.Get("key1"); err != nil || val != "value1" {
		t.Fatal("redis GET option wrong", val, err)
	}

	incr, _ := client.Incr("incr")
	if val, err := client.Incr("incr"); err != nil || incr+1 != val {
		t.Fatal("redis INCR error:", val, err)
	}

	if val, err := client.Get("key2"); err != nil {
		t.Log("redis GET key2:", val)
		t.Fatal("redis SETNX error", err)
	}
	if err := client.SetXX("key2", "setXX", 60*time.Second); err != nil {
		t.Fatal("redis SETXX error", err)
	}
	if err := client.Del("key2"); err != nil {
		t.Fatal("redis DEL error", err)
	}
	if _, err := client.Get("key2"); err == nil {
		t.Fatal("redis DEL error, key: key2")
	}
}
