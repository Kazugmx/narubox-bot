package net.kazugmx

import com.auth0.jwt.JWT
import com.auth0.jwt.algorithms.Algorithm
import io.ktor.http.*
import io.ktor.server.application.*
import io.ktor.server.auth.*
import io.ktor.server.auth.jwt.*
import io.ktor.server.request.*
import io.ktor.server.response.*
import io.ktor.server.routing.*
import net.kazugmx.module.AuthService
import net.kazugmx.module.tryAuth
import net.kazugmx.schema.LoginReq
import net.kazugmx.schema.UserCreateReq
import org.koin.ktor.ext.inject
import java.security.SecureRandom
import java.util.*

fun Application.initAuthUnit() {
    val authService by inject<AuthService>()
    val securePRNG: SecureRandom by inject<SecureRandom>()

    val jwtIssuer = environment.config.property("jwt.issuer").getString()
    val jwtAudience = environment.config.property("jwt.audience").getString()
    val jwtRealm = environment.config.property("jwt.realm").getString()
    val envJwtSecret = environment.config.property("jwt.secret").getString()
    val jwtSecret = if (envJwtSecret == "invalidSecret" || envJwtSecret.isBlank()) {
        log.warn("JWT Secret is not configured! Using temporarily generated secret key.")
        val genToken = ByteArray(32)
        securePRNG.nextBytes(genToken)
        genToken.toHexString()
    } else envJwtSecret


    fun generateToken(id: Int) = JWT.create()
        .withAudience(jwtAudience)
        .withIssuer(jwtIssuer)
        .withClaim("userid", id)
        .withExpiresAt(Date(System.currentTimeMillis() + 60000 * 30 * 4))
        .sign(Algorithm.HMAC256(jwtSecret))


    authentication {
        jwt("auth-jwt") {
            realm = jwtRealm
            verifier(
                JWT
                    .require(Algorithm.HMAC256(jwtSecret))
                    .withAudience(jwtAudience)
                    .withIssuer(jwtIssuer)
                    .build()
            )
            validate { credential ->
                if (credential.payload.audience.contains(jwtAudience)) {
                    if (authService.isUserExists(credential.payload.getClaim("userid").asInt())) {
                        JWTPrincipal(credential.payload)
                    } else null
                } else null
            }
            challenge { _, _ ->
                call.respond(HttpStatusCode.Unauthorized, "Token is not valid or has expired")
            }
        }
    }


    routing {
        route("/api/v1/auth") {
            route("user") {
                authenticate("auth-jwt") {
                    get {
                        call.tryAuth { _, userID ->
                            val selfData = authService.self(
                                userID
                            ) ?: return@tryAuth call.respond(HttpStatusCode.Unauthorized)
                            call.respond(HttpStatusCode.OK, selfData)
                        }
                    }
                }
                post {
                    val req = call.receive<UserCreateReq>()
                    val result = authService.create(req)
                    call.respond(HttpStatusCode.Created, result)
                }
            }
            route("login") {
                get {
                    val token = call.queryParameters["token"] ?: return@get call.respond(HttpStatusCode.BadRequest)
                    val status =authService.verifyCreate(token)
                    if(status){
                        log.info("MailToken is verified.")
                    }else{
                        log.info("MailToken is not verified.")
                        return@get call.respond(HttpStatusCode.Forbidden)
                    }
                    return@get call.respond(HttpStatusCode.OK)
                }

                post {
                    val req = call.receive<LoginReq>()
                    when (val id = authService.login(req)) {
                        //login fail
                        -1 -> return@post call.respond(HttpStatusCode.Unauthorized)
                        //mailToken is not verified
                        -2 -> return@post call.respond(
                            HttpStatusCode.Forbidden,
                            "MailToken is not verified. Please try again."
                        )
                        //login success
                        else -> {
                            val token = generateToken(id)
                            call.response.header(HttpHeaders.Authorization, "Bearer $token")
                            return@post call.respond(
                                HttpStatusCode.OK, mapOf(
                                    "token" to token
                                )
                            )
                        }
                    }
                }
            }
        }
    }
}