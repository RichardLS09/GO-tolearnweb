Redis client For Golang. Based on go-redis
====

Provide a simple way to use go-redis.

- Manager one or more redis client.

How To Use:
```
// create a go file main.go

// import this into your go env

// implement a main method
func main() {
    redis.AddRedis("test", "127.0.0.1:6379", 0)
    
    redisClient := redis.UseRedis("test")
    
    redisClient.Set("mykey", "myvalue", 1 * time.Hour)
    
    if val, err := redisClient.Get("mykey"); err == nil {
        fmt.Println(val)
    } else {
        fmt.Println(err)
        // panic(err)
    }
    
}

```