package main
import (
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