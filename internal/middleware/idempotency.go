package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Simple in-memory cache for idempotency (in production, use Redis)
type IdempotencyCache struct {
	cache map[string]CacheEntry
	mutex sync.RWMutex
}

type CacheEntry struct {
	Response   []byte
	StatusCode int
	Headers    map[string]string
	Timestamp  time.Time
}

var globalCache = &IdempotencyCache{
	cache: make(map[string]CacheEntry),
}

// IdempotencyMiddleware provides idempotency for POST requests
func IdempotencyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only apply to POST requests (create operations)
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}

		// Check for idempotency key header
		idempotencyKey := r.Header.Get("Idempotency-Key")
		if idempotencyKey == "" {
			// If no idempotency key, continue without caching
			next.ServeHTTP(w, r)
			return
		}

		// Create a unique key based on the request
		requestKey, err := createRequestKey(r, idempotencyKey)
		if err != nil {
			http.Error(w, "Failed to process idempotency key", http.StatusInternalServerError)
			return
		}

		// Check if we've seen this request before
		if cachedResponse, found := getCachedResponse(requestKey); found {
			// Return cached response
			for key, value := range cachedResponse.Headers {
				w.Header().Set(key, value)
			}
			w.WriteHeader(cachedResponse.StatusCode)
			w.Write(cachedResponse.Response)
			return
		}

		// Capture the response
		responseWriter := &ResponseCapture{
			ResponseWriter: w,
			body:           make([]byte, 0),
			headers:        make(map[string]string),
		}

		next.ServeHTTP(responseWriter, r)

		// Cache the response for future requests (only if successful)
		if responseWriter.statusCode >= 200 && responseWriter.statusCode < 300 {
			cacheResponse(requestKey, CacheEntry{
				Response:   responseWriter.body,
				StatusCode: responseWriter.statusCode,
				Headers:    responseWriter.headers,
				Timestamp:  time.Now(),
			})
		}
	})
}

// ResponseCapture captures the response for caching
type ResponseCapture struct {
	http.ResponseWriter
	body       []byte
	statusCode int
	headers    map[string]string
}

func (rc *ResponseCapture) Write(data []byte) (int, error) {
	rc.body = append(rc.body, data...)
	return rc.ResponseWriter.Write(data)
}

func (rc *ResponseCapture) WriteHeader(statusCode int) {
	rc.statusCode = statusCode
	rc.ResponseWriter.WriteHeader(statusCode)
}

func (rc *ResponseCapture) Header() http.Header {
	headers := rc.ResponseWriter.Header()
	// Capture important headers
	for key, values := range headers {
		if len(values) > 0 {
			rc.headers[key] = values[0]
		}
	}
	return headers
}

// createRequestKey creates a unique key for the request
func createRequestKey(r *http.Request, idempotencyKey string) (string, error) {
	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}

	// Restore the body for the next handler
	r.Body = io.NopCloser(strings.NewReader(string(body)))

	// Create hash of method + path + body + idempotency key
	hasher := sha256.New()
	hasher.Write([]byte(r.Method))
	hasher.Write([]byte(r.URL.Path))
	hasher.Write(body)
	hasher.Write([]byte(idempotencyKey))

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// getCachedResponse retrieves a cached response if it exists and is still valid
func getCachedResponse(key string) (CacheEntry, bool) {
	globalCache.mutex.RLock()
	defer globalCache.mutex.RUnlock()

	entry, found := globalCache.cache[key]
	if !found {
		return CacheEntry{}, false
	}

	// Check if entry is still valid (24 hours)
	if time.Since(entry.Timestamp) > 24*time.Hour {
		// Entry is too old, remove it
		delete(globalCache.cache, key)
		return CacheEntry{}, false
	}

	return entry, true
}

// cacheResponse stores a response in the cache
func cacheResponse(key string, entry CacheEntry) {
	globalCache.mutex.Lock()
	defer globalCache.mutex.Unlock()

	globalCache.cache[key] = entry

	// Simple cleanup: if cache gets too large, remove old entries
	if len(globalCache.cache) > 10000 {
		cleanupOldEntries()
	}
}

// cleanupOldEntries removes entries older than 1 hour
func cleanupOldEntries() {
	cutoff := time.Now().Add(-1 * time.Hour)
	for key, entry := range globalCache.cache {
		if entry.Timestamp.Before(cutoff) {
			delete(globalCache.cache, key)
		}
	}
}
