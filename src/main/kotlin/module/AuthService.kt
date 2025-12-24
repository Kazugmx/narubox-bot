package net.kazugmx.module

import at.favre.lib.crypto.bcrypt.BCrypt
import com.auth0.jwt.exceptions.JWTDecodeException
import com.auth0.jwt.exceptions.TokenExpiredException
import io.ktor.http.*
import io.ktor.server.application.*
import io.ktor.server.auth.*
import io.ktor.server.auth.jwt.*
import io.ktor.server.response.*
import kotlinx.coroutines.Dispatchers
import net.kazugmx.schema.*
import org.jetbrains.exposed.sql.*
import org.jetbrains.exposed.sql.SqlExpressionBuilder.eq
import org.jetbrains.exposed.sql.kotlin.datetime.CurrentDateTime
import org.jetbrains.exposed.sql.transactions.experimental.newSuspendedTransaction
import org.koin.ktor.ext.inject
import org.slf4j.Logger
import java.security.SecureRandom
import kotlin.getValue


class AuthService(
    @Suppress("unused") db: Database,
    private val isMailActive: Boolean,
    private val origin: String,
    private val mailConfig: MailConfig,
    private val logger: Logger,
    private val newUser: Boolean,
    private val securePRNG: SecureRandom
) {
    init {
        logger.info("initialized")
        if (!isMailActive) logger.info("Mail is disabled.")
    }

    val mailClient: MailClient? = if (isMailActive) MailClient(mailConfig) else null

    companion object {
        private const val MAX_ATTEMPTS = 5
        private const val LOCK_TIME_MS = 5 * 60 * 1000L // 5分
    }

    private fun updateTime(id: Int) {
        UserTable.update(where = { UserTable.id eq id }) {
            it[lastAccessAt] = CurrentDateTime
        }
    }


    suspend fun create(req: UserCreateReq): UserCreateRes {
        if (!newUser) return UserCreateRes(status = "user registration is disabled.")
        val pwHash = BCrypt.withDefaults().hashToString(16, req.password.toCharArray())
        var existUser: ResultRow? = null
        dbQuery {
            existUser = UserTable.select(
                UserTable.username
            ).where {
                (UserTable.username eq req.username) or (UserTable.mail eq req.mail)
            }.singleOrNull()
        }
        if (existUser != null) {
            return UserCreateRes(status = "failed")
        }
        try {
            var mailToken: String? = null
            if (isMailActive && mailClient != null) {
                val genToken = ByteArray(32)
                securePRNG.nextBytes(genToken)
                mailToken = genToken.toHexString()

                mailClient.send(
                    req.mail,
                    "Mail verification - narubox-bot",
                    """
                        Your account has been created.
                        This is your URL for verification: https://${origin}/api/v1/auth/login?token=${mailToken}
                        
                        If it doesn't work, please contact the administrator.
                    """.trimIndent()
                )
            }
            dbQuery {
                UserTable.insert {
                    it[username] = req.username
                    it[mail] = req.mail
                    it[password] = pwHash
                    it[UserTable.mailToken] = mailToken
                }[UserTable.id].value
            }
            return UserCreateRes(status = "success")
        } catch (e: Exception) {
            logger.error("error while creating user", e)
            return UserCreateRes(status = "failed")
        }

    }

    suspend fun verifyCreate(token: String) = dbQuery {
        val user = UserTable.select(
            UserTable.id,
            UserTable.mailToken,
            UserTable.mail).where { UserTable.mailToken eq token }
            .singleOrNull()
            ?: return@dbQuery false

        if (
            token == user[UserTable.mailToken]
        ) {
            UserTable.update({ UserTable.id eq user[UserTable.id] }) {
                it[mailToken] = null
            }
            mailClient?.send(
                user[UserTable.mail],
                "Welcome to narubox-bot!",
                "Thank you for verifying your account. You can now log in to your account."
            )
            return@dbQuery true
        } else return@dbQuery false
    }

    private val loginAttempts = mutableMapOf<String, Pair<Int, Long>>()

    suspend fun login(loginReq: LoginReq): Int = dbQuery {
        val username = loginReq.username
        val now = System.currentTimeMillis()

        val (attempts, unlockTime) =
            loginAttempts[username] ?: (0 to 0L)

        if (now < unlockTime) {
            logger.warn("Login locked: {}", username)
            return@dbQuery -1
        }

        //login process

        @Suppress("LocalVariableName")
        val FALLBACK_HASH = $$"$2b$16$C6UzMDM.H6dfI/f/IKcCcO4uP04Jw8A61uYyYV3D1h0WyZxWj96C2"
        val challUserData = UserTable
            .select(UserTable.id, UserTable.password, UserTable.mailToken, UserTable.createdAt)
            .where { UserTable.username eq username }
            .singleOrNull()
        val targetHash = challUserData?.get(UserTable.password)
            ?: FALLBACK_HASH

        val verified = BCrypt
            .verifyer().verify(
                loginReq.password.toCharArray(),
                targetHash.toCharArray()
            ).verified

        if (challUserData?.get(UserTable.mailToken) != null) {
            return@dbQuery -2
        }

        if (verified && challUserData != null) {
            val intID = challUserData[UserTable.id].value
            logger.info("userid:{} is logged in.", intID)
            updateTime(intID)
            loginAttempts.remove(username)
            return@dbQuery intID
        }

        //rate limit per user

        val newAttempts = attempts + 1
        if (newAttempts >= MAX_ATTEMPTS) {
            loginAttempts[username] = (newAttempts to (now + LOCK_TIME_MS))
            logger.warn("Login locked: {}", username)
        } else {
            loginAttempts[username] = (newAttempts to 0L)
            logger.warn("Login failed: {} (attempt {}/{})", username, newAttempts, MAX_ATTEMPTS)
        }
        logger.info("failed to login.")
        return@dbQuery -1
    }


    suspend fun self(id: Int): SelfRes? = dbQuery {
        updateTime(id)
        UserTable.select(
            UserTable.mail, UserTable.username,
            UserTable.createdAt, UserTable.lastAccessAt
        ).where(UserTable.id eq id).singleOrNull(
        )?.let {
            SelfRes(
                mail = it[UserTable.mail],
                username = it[UserTable.username],
                createdAt = it[UserTable.createdAt],
                lastAccess = it[UserTable.lastAccessAt]
            )
        }
    }

    suspend fun isUserExists(id: Int): Boolean = dbQuery {
        updateTime(id)
        UserTable.select(UserTable.id).where { UserTable.id eq id }.singleOrNull()
            ?: return@dbQuery false
        return@dbQuery true
    }

    private suspend fun <T> dbQuery(block: suspend () -> T): T =
        newSuspendedTransaction(Dispatchers.IO) { block() }
}

suspend inline fun ApplicationCall.tryAuth(block: (JWTPrincipal, Int) -> Unit) {
    try {
        val principal =
            principal<JWTPrincipal>() ?: return respond(
                HttpStatusCode.BadRequest,
                mapOf("status" to "failed", "reason" to "No JWT Principal")
            )
        val userID = principal.payload.getClaim("userid").asInt()
        block(principal, userID)
    } catch (e: TokenExpiredException) {
        respond(
            HttpStatusCode.BadRequest,
            mapOf(
                "status" to "failed",
                "reason" to "Token Expired",
                "expiredOn" to e.expiredOn.toString(),
            ),
        )
        application.log.info("Token Expired: ${e.expiredOn}")
    } catch (e: JWTDecodeException) {
        respond(
            HttpStatusCode.BadRequest,
            mapOf(
                "status" to "failed",
                "reason" to "Invalid Token",
                "detail" to e.message,
            ),
        )
    } catch (e: Exception) {
        application.log.error("JWT processing failed", e)
        respond(
            HttpStatusCode.InternalServerError,
            mapOf(
                "status" to "failed",
                "reason" to "Internal Server Error",
            )
        )
    }
}
