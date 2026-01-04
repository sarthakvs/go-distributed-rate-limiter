# 🛡️ AegisAPI - Distributed Rate Limiter

A high-performance, distributed API rate-limiting middleware built with **Go** and **Redis**.

<img width="620" height="115" alt="image" src="https://github.com/user-attachments/assets/694a9db4-9e91-4dec-8cb6-40d3eb534c87" />

In case you're rate-limited:

<img width="632" height="164" alt="image" src="https://github.com/user-attachments/assets/3a157a18-8a1d-4e26-811e-373aaf161801" />


## 🚀 Overview
AegisAPI implements the **Token Bucket Algorithm** to protect your services from abuse and traffic spikes. Unlike standard in-memory limiters, AegisAPI is designed for distributed systems where multiple server instances need to share a global state.

## ✨ Features
- **Phase 1 (In-Memory Logic):** High-concurrency support using Go's `sync.Mutex`.
- **Phase 2 (Distributed Logic):** Atomic **Lua scripts** executed within Redis to ensure consistency across multiple nodes without race conditions.
- **Fail-Open Strategy:** Designed to prioritize uptime; if Redis experiences latency, the system allows traffic to ensure service continuity.
- **Cloud Native:** Configured for easy deployment on **Render** and **Redis Cloud**.

## 🛠️ Tech Stack
- **Language:** Go (Golang)
- **Database:** Redis (Lua scripting for atomicity)
- **Deployment:** Render

## ⚙️ Environment Variables
To run this project in production, set the following environment variables:
- `REDIS_ADDR`: Your Redis host and port (e.g., `redis-123.redislabs.com:18273`)
- `REDIS_USERNAME`: Your Redis username
- `REDIS_PASSWORD`: Your Redis password
- `PORT`: The port your Go server will listen on (default: `8080`)

## 🚦 Usage
Once the server is running, hit the `/ping` endpoint with a `user` query parameter:
`GET /ping?user=sarthakvs`

## 📝 Implementation Details
- **Algorithm:** Token Bucket
- **Precision:** Second-based (Lazy refill)
- **Concurrency:** Thread-safe atomic operations via Redis Lua

## 💻 Local Development
1. Clone the repository: `git clone https://github.com/sarthakvs/go-distributed-rate-limiter.git`
2. Install dependencies: `go mod download`
3. Set your environment variables (or use a `.env` file).
4. Run the server: `go run phase2redis.go` (you can run the `phase1inMemory.go` without setting the enviroment variables
