package net.kazugmx

import io.ktor.http.*
import io.ktor.server.application.*
import io.ktor.server.auth.*
import io.ktor.server.request.*
import io.ktor.server.response.*
import io.ktor.server.routing.*
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.launch
import kotlinx.serialization.decodeFromString
import net.kazugmx.module.BotService
import net.kazugmx.module.tryAuth
import net.kazugmx.schema.BotRegisterReq
import net.kazugmx.schema.Feed
import net.kazugmx.schema.SubscribeRequest
import nl.adaptivity.xmlutil.ExperimentalXmlUtilApi
import nl.adaptivity.xmlutil.serialization.DefaultXmlSerializationPolicy
import nl.adaptivity.xmlutil.serialization.UnknownChildHandler
import nl.adaptivity.xmlutil.serialization.XML
import org.koin.ktor.ext.inject

@OptIn(ExperimentalXmlUtilApi::class)
fun Application.initBotUnit() {
    val bot by inject<BotService>()
    val xml = XML {
        autoPolymorphic = false
        policy = DefaultXmlSerializationPolicy.Builder().apply {
            pedantic = false
            unknownChildHandler = UnknownChildHandler({ _, _, _, _, _ -> emptyList() })
        }.build()
    }
    val backScope = CoroutineScope(Dispatchers.IO + SupervisorJob())

    fun parseFeed(payload: String): Feed? = try {
        xml.decodeFromString(payload)
    } catch (e: Exception) {
        log.error("Failed to parse XML feed: {}", e.message)
        null
    }

    routing {
        route("/api/v1/bot") {
            route("pubsub/{endpointID}"){
                get {
                    val chall = call.queryParameters["hub.challenge"]
                    if (chall != null) {
                        log.info("PubSub challenge: {} / endpoint {}", chall, call.parameters["endpointID"])
                        call.respondText(chall)
                    } else call.respond(HttpStatusCode.BadRequest)
                }
                post {
                    val endpoint = call.parameters["endpointID"] ?: return@post call.respond(HttpStatusCode.OK)
                    val body = call.receiveText()
                    val feed = parseFeed(body) ?: return@post call.respond(HttpStatusCode.OK)
                    val isDeletedEntry = body.contains("deleted-entry")
                    if (isDeletedEntry) {
                        log.info(
                            "Deleted entry passed. Skipped notifying."
                        )
                    }
                    log.info("PubSub body: \n{}", body)

                    if (isDeletedEntry) return@post call.respond(HttpStatusCode.OK)

                    feed.entry?.forEach {
                        log.info("channelID {} / https://www.youtube.com/watch?v={}", it.channelID, it.videoID)
                        it.videoID?.let { videoID -> backScope.launch { bot.notifyToBots(videoID = videoID,endpointID=endpoint) } }
                    }
                    call.respond(HttpStatusCode.OK)
                }
            }


            authenticate("auth-jwt") {
                get("force") {
                    bot.refreshChannel()
                    call.respond(HttpStatusCode.OK)
                }
                get {
                    call.tryAuth { principal, _ ->
                        val botList = bot.getBots(principal.payload.getClaim("userid").asInt())
                        call.respond(HttpStatusCode.OK, botList)
                    }
                }
                post("register") {
                    call.tryAuth { _, userID ->
                        val req = call.receive<BotRegisterReq>()
                        val botRegister = bot.registerBot(
                            req, userID
                        )
                        if (botRegister.success) {
                            call.respond(HttpStatusCode.Accepted, botRegister)
                        } else call.respond(HttpStatusCode.BadRequest, botRegister)
                    }
                }
                route("{botID}") {
                    get {
                        call.tryAuth { _, userID ->
                            val botID =
                                call.parameters["botID"] ?: return@tryAuth call.respond(HttpStatusCode.BadRequest)
                            val channels = bot.getChannels(botID, userID)
                            call.respond(HttpStatusCode.OK, mapOf("channels" to channels))
                        }
                    }
                    delete {
                        call.tryAuth { _, userID ->
                            val botID =
                                call.parameters["botID"] ?: return@tryAuth call.respond(HttpStatusCode.BadRequest)
                            bot.unregisterBot(botID, userID)
                            call.respond(HttpStatusCode.Accepted, mapOf("success" to true))
                        }
                    }
                    post {
                        call.tryAuth { _, userID ->
                            val req = call.receive<SubscribeRequest>()
                            val botID =
                                call.parameters["botID"] ?: return@tryAuth call.respond(HttpStatusCode.BadRequest)

                            if (bot.registerChannel(botID, req, userID)) {
                                call.respond(HttpStatusCode.Accepted, mapOf("success" to true))
                            }
                            call.respond(HttpStatusCode.BadRequest, mapOf("success" to false))
                        }
                    }
                    delete("channel/{channelID}") {
                        call.tryAuth { _, userID ->
                            val botID =
                                call.parameters["botID"] ?: return@tryAuth call.respond(HttpStatusCode.BadRequest)
                            val channelID =
                                call.parameters["channelID"] ?: return@tryAuth call.respond(HttpStatusCode.BadRequest)

                            bot.deleteChannel(userID, botID, channelID)
                            call.respond(HttpStatusCode.Accepted, mapOf("success" to true))
                        }
                    }
                }
            }
        }
    }
}
