package net.kazugmx.module.util

import java.util.concurrent.ConcurrentHashMap

class LoginAttemptStore(
    private val maxAttempts: Int,
    private val lockTimeMs: Long,
    private val maxEntries: Int = 50_000, // 上限（運用に合わせて）
) {
    private data class Entry(val attempts: Int, val unlockAt: Long, val lastSeen: Long)
    private val map = ConcurrentHashMap<String, Entry>()

    fun checkLocked(key: String, now: Long = System.currentTimeMillis()): Long? {
        val e = map[key] ?: return null
        if (now < e.unlockAt) return e.unlockAt
        return null
    }

    fun recordFailure(key: String, now: Long = System.currentTimeMillis()): Pair<Int, Long> {
        // ざっくり上限対策（厳密LRUでなくても実用上効く）
        if (map.size > maxEntries) {
            map.clear()
        }

        val old = map[key]
        val nextAttempts = (old?.attempts ?: 0) + 1
        val unlockAt = if (nextAttempts >= maxAttempts) now + lockTimeMs else 0L
        map[key] = Entry(nextAttempts, unlockAt, now)
        return nextAttempts to unlockAt
    }

    fun reset(key: String) {
        map.remove(key)
    }
}
