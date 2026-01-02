package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type Bucket struct {
	capacity 	float64  // to compare easily with tokens, we have it as float
	tokens 		float64  // so everytime a request comes the timediff is added 
	reRate 		float64 // the division result can be a float
	lastReTime 	time.Time // time object
	mu 			sync.Mutex // to have a lock mechanism
}

func (b *Bucket) isAllowed() bool { //function to check if the request can be allowed or not 
	
	b.mu.Lock() // if 2 request comes in and goroutine executes both at the same nanosecond we need to have a lock
	defer b.mu.Unlock()

	now:=time.Now() // getting the curr time 
	timediff := now.Sub(b.lastReTime).Seconds() // time difference since the last refill 
	b.tokens += timediff * b.reRate // adding the tokens
	if b.tokens > b.capacity {
		b.tokens = b.capacity // in case of overflow
	}
	b.lastReTime = now // setting the curr time

	if b.tokens>=1.0{ //check to see if we have an entire token to serve one single request
		b.tokens -= 1.0 
		return true
	}

	return false
}
 // we created a single bucket, now many users will have their own buckets, so we will have a map of buckets
type RateLimiter struct{
	buckets map[string]*Bucket // key is userID and the value is bucket address so the changes done to the bucket persist
	mu 		sync.Mutex // here too we will have a lock so that no 2 user access the map at the same time
	rate 	float64 // all buckets have the same rate at first
	cap 	float64 // all buckets have the same capacity at first
}

// creating a new limiter 
func newLimiter(rate float64,cap float64) *RateLimiter{ 
	return &RateLimiter{
		buckets: make(map[string]*Bucket), 
		rate: rate,
		cap: cap,
	}
}
// getting the bucket based on the userID (or creating it)
func (r *RateLimiter) getBucket(userID string)*Bucket{ 
	r.mu.Lock() // lock the map for a few microseconds
	defer r.mu.Unlock()

	bucket,exists := r.buckets[userID] // retrieving the bucket
	if !exists { // if bucket does not exist, creating one
		bucket = &Bucket{
			capacity: r.cap,
			tokens: r.cap,
			reRate: r.rate,
			lastReTime: time.Now(),
		}
		r.buckets[userID] = bucket 
	}
	return bucket
} 
//middleware to check if the request is valid or no
func (rl *RateLimiter) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		userID := r.URL.Query().Get("user")
		if userID==""{
			http.Error(w,"User ID required",http.StatusBadRequest)
			return 
		}
		bucket:= rl.getBucket(userID)
		if bucket.isAllowed(){
			next.ServeHTTP(w,r)
		} else{
			http.Error(w,"Too many requests! Try again after some time.", http.StatusTooManyRequests)
		}
	})
}

func main(){
	limiter:=newLimiter(0.2,3)
	helloHandler:= http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		fmt.Fprintln(w,"Hello! Your request was successful")
	})
	http.Handle("/ping",limiter.middleware(helloHandler))
	port := os.Getenv("PORT")
	if port==""{
		port="3000"
	}
	fmt.Println("Server started at port",port)
	http.ListenAndServe("0.0.0.0:"+port,nil)
}