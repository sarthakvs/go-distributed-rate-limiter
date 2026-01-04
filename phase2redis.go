package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"github.com/redis/go-redis/v9"
)
var ctx = context.Background()
type RedisLimiter struct {
	client 		*redis.Client
	rate		float64
	capacity 	float64
}

func newRedisLimiter (rate,capacity float64) *RedisLimiter{
	addr := os.Getenv("REDIS_ADDR")
	pass := os.Getenv("REDIS_PASSWORD")
	user := os.Getenv("REDIS_USERNAME")
	if addr=="" || pass=="" || user=="" {
		fmt.Println("Missing required Redis environment variables")
		return nil
	}
	rdb :=redis.NewClient(&redis.Options{
		Addr: addr,
		Username: user,
		Password: pass,
		DB: 0,
	})
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}
	return &RedisLimiter{
		client: rdb,
		rate: rate,
		capacity: capacity,
	}
}
func(rl *RedisLimiter) isAllowed(userID string) bool{
	//in phase 1 we require the lock, but here redis takes care of that
	// on its own, the lua script is atomic, redis won't run any other
	//command until this finishes
	luaScript := `
			local key = KEYS[1] --userID
			local cap = tonumber(ARGV[1]) --bucketsize (in lua indexes start from 1)
			local rate = tonumber(ARGV[2]) -- refill rate
			local now = tonumber(ARGV[3]) -- current timestamp in go

			local bucket = redis.call("HMGET",key,"tokens","last_time")
			
			local tokens = tonumber(bucket[1]) or cap
			local last_time = tonumber(bucket[2]) or now
			--1. Refill logic: calculate how much has passed and then the tokens
			local delta = math.max(0,now-last_time) 
			tokens = math.min(cap,tokens + (delta*rate))
			--2. consumption logic: can we afford 1 singular request?
			local allowed = 0
			if tokens>=1 then
				tokens = tokens - 1
				allowed = 1 --1 means 'true'
			end
			--3. Update redis: save the token count back to the redis
			redis.call("HMSET",key,"tokens",tokens,"last_time",now)

			--4. cleanup: set the TTL of this hash
			redis.call("EXPIRE",key,60)

			return allowed
	`
	redisKey := "userKey:" + userID
	result,err := rl.client.Eval(ctx,luaScript,[]string{redisKey},rl.capacity,rl.rate,time.Now().Unix()).Int()
	if err != nil {
		fmt.Printf("Redis error for %s: %v\n",userID,err)
		return true
	}
	return result==1
}

func (rl *RedisLimiter) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		userID:= r.URL.Query().Get("user")
		if userID==""{
			http.Error(w,"User ID required (?user=XYZ)",http.StatusBadRequest)
			return 
		}
		if rl.isAllowed(userID){
			next.ServeHTTP(w,r)
		} else{
			w.Header().Set("Content-Type","application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprintln(w,`{"error":"Too many requests. Slow down!"}`)
		}
	})
}

func main(){
	limiter := newRedisLimiter(0.2,4) // 5 second = 1 req and 4 tokens already
	helloHandler:= http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		fmt.Fprintf(w,"Hello! Your request was successful. Time : %v\n",time.Now().Format(time.RFC3339))
	})
	http.Handle("/ping",limiter.middleware(helloHandler))
	port := os.Getenv("PORT")
	if port==""{
		port="8080"
	}
	fmt.Println("Server started at port:",port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port,nil))
}


