package net.kazugmx.schema

import kotlinx.datetime.LocalDateTime
import kotlinx.serialization.Serializable
import org.jetbrains.exposed.dao.id.IntIdTable
import org.jetbrains.exposed.sql.kotlin.datetime.CurrentDateTime
import org.jetbrains.exposed.sql.kotlin.datetime.datetime

// user creation payloads
@Serializable
data class UserCreateReq(
    val username: String,
    val mail: String,
    val password: String,
)

@Serializable
data class UserCreateRes(
    val status: String,
    val datetime: LocalDateTime? = null,
)

// login payloads

@Serializable
data class LoginReq(
    val username: String,
    val password: String,
)

@Serializable
data class SelfRes(
    val mail: String,
    val username: String,
    val createdAt: LocalDateTime,
    val lastAccess: LocalDateTime? = null
)

object UserTable : IntIdTable("user_table") {
    val username = varchar("username", 50).uniqueIndex()
    val mail = varchar("mail", length = 255).uniqueIndex()
    val password = varchar("password", 100)
    val createdAt = datetime("created_at").defaultExpression(CurrentDateTime)
    val lastAccessAt = datetime("last_access").nullable()
}