A reliable and thread-safe rate limiting library with Redis as backend


It’s easy to develop a rate limiter for your app, use mutex or other mechanisms to avoid eventual race conditions. However, these mechanisms work well when you have a single copy of the app running. In the current generation of computing, high availability is becoming defacto standard for app development. Keeping this in mind, it’s prudent to redesign the rate limiting logic as the old logic of mitigation of race condition won’t work with multiple copies of the same app running, possibly in different machines/containers.

The ‘ratelimit’ library uses Redis as backend (you need to install Redis separately) store to manage keys and their respective counter values and stop the counter before it can cross the threshold value. Redis is fast and in-memory based key-value store and operations on Redis are thread-safe.

Refer to example.go for a live example..
