package middleware

import (
    "net/http"
    "sync"
    "time"
    "github.com/gin-gonic/gin"
)

var (
    visitors = make(map[string]*visitor)
    mu       sync.Mutex
)

type visitor struct {
    count    int
    lastSeen time.Time
}

func cleanupVisitors() {
    for {
        time.Sleep(3 * time.Minute)
        mu.Lock()
        for ip, v := range visitors {
            if time.Since(v.lastSeen) > 3*time.Minute {
                delete(visitors, ip)
            }
        }
        mu.Unlock()
    }
}

func RateLimit(maxPerMinute int) gin.HandlerFunc {
    go cleanupVisitors()
    return func(c *gin.Context) {
        ip := c.ClientIP()
        mu.Lock()
        v, exists := visitors[ip]
        if !exists || time.Since(v.lastSeen) > time.Minute {
            visitors[ip] = &visitor{count: 1, lastSeen: time.Now()}
            mu.Unlock()
            c.Next()
            return
        }
        
        v.count++
        if v.count > maxPerMinute {
            mu.Unlock()
            c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
                "error": "Rate limit exceeded",
            })
            return
        }
        mu.Unlock()
        c.Next()
    }
}
