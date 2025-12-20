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
import java.util.*

fun Application.initAuthUnit(authService: AuthService) {
    val jwtIssuer = environment.config.property("jwt.issuer").getString()
    val jwtAudience = environment.config.property("jwt.audience").getString()
    val jwtRealm = environment.config.property("jwt.realm").getString()
    val jwtSecret = environment.config.property("jwt.secret").getString()
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
                post {
                    val req = call.receive<LoginReq>()
                    val id = authService.login(req)
                    if (id != -1) {
                        val token = generateToken(id)
                        call.response.header(HttpHeaders.Authorization, "Bearer $token")
                        return@post call.respond(
                            HttpStatusCode.OK, mapOf(
                                "token" to token
                            )
                        )
                    } else return@post call.respond(HttpStatusCode.Unauthorized)
                }
            }
        }
    }
}