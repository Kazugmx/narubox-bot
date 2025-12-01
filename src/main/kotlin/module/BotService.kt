package net.kazugmx.module

import io.ktor.client.*
import io.ktor.client.call.*
import io.ktor.client.engine.cio.*
import io.ktor.client.plugins.contentnegotiation.*
import io.ktor.client.request.*
import io.ktor.http.*
import io.ktor.serialization.kotlinx.json.*
import kotlinx.coroutines.Dispatchers
import kotlinx.serialization.json.Json
import net.kazugmx.schema.*
import org.jetbrains.exposed.sql.*
import org.jetbrains.exposed.sql.SqlExpressionBuilder.eq
import org.jetbrains.exposed.sql.kotlin.datetime.CurrentDateTime
import org.jetbrains.exposed.sql.transactions.experimental.newSuspendedTransaction
import org.jetbrains.exposed.sql.transactions.transaction
import org.slf4j.Logger
import java.time.Instant
import java.time.OffsetDateTime
import java.util.*
import kotlin.uuid.ExperimentalUuidApi


@OptIn(ExperimentalUuidApi::class)
class BotService(
    @Suppress("unused") db: Database,
    private val logger: Logger,
    private val apiKey: String,
    private val origin: String
) {
    init {
        transaction {
            SchemaUtils.create(ChannelTable)
            SchemaUtils.create(ChannelRegTable)
            SchemaUtils.create(BotRegTable)
            SchemaUtils.create(OnAirTable)
            logger.info("initialized")
        }
    }

    companion object {
        private val client = HttpClient(CIO) {
            install(ContentNegotiation) {
                json(
                    Json {
                        ignoreUnknownKeys = true
                        isLenient = true
                    }
                )
            }
        }
    }

    private suspend fun <T> dbQuery(block: suspend () -> T): T =
        newSuspendedTransaction(Dispatchers.IO) { block() }

    private fun convertTimeEpochSec(time: String): Long = OffsetDateTime.parse(time).toEpochSecond()

    private suspend fun classifyStatusToNotify(data: YtResponse.QueryData): DeliverState = dbQuery {
        var state: DeliverState = DeliverState.ERROR

        val idExist = OnAirTable.select(OnAirTable.videoID, OnAirTable.previousState)
            .where { OnAirTable.videoID eq data.videoId }.singleOrNull()
        if (idExist != null) {
            //handle duplicate for live streams
            if (idExist[OnAirTable.previousState] != data.type) {
                if (
                    idExist[OnAirTable.previousState] == DeliverState.ON_AIR.text &&
                    idExist[OnAirTable.previousState] != data.type
                ) {
                    OnAirTable.update( { OnAirTable.videoID eq idExist[OnAirTable.videoID] }){
                        it[OnAirTable.previousState] = DeliverState.VIDEO.text
                    }
                    state = DeliverState.FINISHED
                } else if (
                    idExist[OnAirTable.previousState] == DeliverState.UPCOMING.text &&
                    idExist[OnAirTable.previousState] != data.type
                ) {
                    OnAirTable.update({ OnAirTable.videoID eq idExist[OnAirTable.videoID] }) {
                        it[OnAirTable.previousState] = DeliverState.ON_AIR.text
                    }
                    state = DeliverState.ON_AIR
                }
            }
        } else {
            OnAirTable.insert {
                it[OnAirTable.videoID] = data.videoId
                it[OnAirTable.previousState] = data.type
            }
            when (data.type) {
                DeliverState.ON_AIR.text -> {
                    state = DeliverState.ON_AIR
                }
                DeliverState.UPCOMING.text -> {
                    state = DeliverState.UPCOMING
                }
                DeliverState.VIDEO.text -> {
                    state = DeliverState.VIDEO
                }
            }
        }
        return@dbQuery state
    }

    suspend fun getBots(userID: Int) = dbQuery {
        BotRegTable.select(
            BotRegTable.id, BotRegTable.label, BotRegTable.wsUrl, BotRegTable.mentionRoleID
        ).where { BotRegTable.ownerID eq userID }.map {
            BotListRes(
                botID = it[BotRegTable.id].value.toString(),
                label = it[BotRegTable.label],
                wsUrl = it[BotRegTable.wsUrl],
                mentionRoleID = it[BotRegTable.mentionRoleID]
            )
        }
    }

    suspend fun registerBot(request: BotRegisterReq, userID: Int): BotRegisterRes = dbQuery {
        val hasConflict = BotRegTable.select(
            BotRegTable.id
        ).where {
            (BotRegTable.label eq request.botLabel) and
                    (BotRegTable.ownerID eq userID)
        }.singleOrNull() != null
        if (hasConflict) return@dbQuery BotRegisterRes()
        val botID = BotRegTable.insert {
            it[BotRegTable.label] = request.botLabel
            it[BotRegTable.wsUrl] = request.wsUrl
            it[BotRegTable.ownerID] = userID
            it[BotRegTable.mentionRoleID] = request.mentionRoleID
        }[BotRegTable.id].value.toString()

        val mentionIDPlacer = if (request.mentionRoleID.contains("ignore:")){
            request.mentionRoleID.removePrefix("ignore:")
        } else "<@&${request.mentionRoleID}>"

        client.post(request.wsUrl) {
            contentType(ContentType.Application.Json)
            setBody(mapOf("content" to "Hello world, $mentionIDPlacer !"))
        }
        return@dbQuery BotRegisterRes(true, botID)
    }

    suspend fun unregisterBot(botID: String, userID: Int) = dbQuery {
        BotRegTable.deleteWhere {
            logger.info("Bot {} unregistered by user {}", botID, userID)
            BotRegTable.id eq UUID.fromString(botID) and
                    (BotRegTable.ownerID eq userID)
        }
    }

    suspend fun getChannels(botID: String, userID: Int): List<String> = dbQuery {
        @Suppress("LocalVariableName")
        val botID_u = UUID.fromString(botID)
        val owner = BotRegTable.select(BotRegTable.ownerID).where {
            (BotRegTable.id eq botID_u) and (BotRegTable.ownerID eq userID)
        }
            .singleOrNull()
        if (owner == null) return@dbQuery emptyList()
        ChannelRegTable.select(ChannelRegTable.channelID)
            .where {
                (ChannelRegTable.botID eq botID_u)
            }.map { it[ChannelRegTable.channelID] }
    }

    suspend fun registerChannel(botID: String, subReq: SubscribeRequest, userID: Int = -1): Boolean {

        val botID = UUID.fromString(botID)
        val channelID = subReq.channelID
        val isOwner = dbQuery {
            BotRegTable.select(BotRegTable.ownerID).where {
                (BotRegTable.id eq botID) and (BotRegTable.ownerID eq userID)
            }.singleOrNull()?.let { return@dbQuery true }
            return@dbQuery false
        }
        if (!isOwner) return false

        val hasRegisteredOnChannels = dbQuery {
            ChannelTable.select(ChannelTable.channelID).where {
                ChannelTable.channelID eq channelID
            }.singleOrNull()?.let { true } ?: false
        }

        val hasRegisteredOnBot = dbQuery {
            ChannelRegTable.select(ChannelRegTable.botID).where {
                (ChannelRegTable.botID eq botID) and (ChannelRegTable.channelID eq channelID)
            }.singleOrNull()?.let { true } ?: false
        }

        if (!hasRegisteredOnChannels || subReq.refresh) {
            val hubUrl = "https://pubsubhubbub.appspot.com/subscribe"
            val topicUrl = "https://www.youtube.com/xml/feeds/videos.xml?channel_id=$channelID"
            val callbackUrl = "https://${origin}/api/v1/bot/pubsub"

            val a = client.post(hubUrl) {
                url {
                    parameters.append("hub.callback", callbackUrl)
                    parameters.append("hub.lease_seconds", "864000")
                    parameters.append("hub.mode", "subscribe")
                    parameters.append("hub.topic", topicUrl)
                    parameters.append("hub.verify", "sync")
                }
            }

            dbQuery {
                if (!hasRegisteredOnBot) ChannelTable.insert { it[ChannelTable.channelID] = channelID }
                else ChannelTable.update({ ChannelTable.channelID eq channelID }) {
                    it[ChannelTable.lastUpdate] = CurrentDateTime
                }
            }

            logger.info(
                "status: {} / Subscribed to channel {} by bot {}",
                a.status.value, channelID, botID
            )
        }
        if (!hasRegisteredOnBot) dbQuery {
            ChannelRegTable.insert {
                it[ChannelRegTable.botID] = botID
                it[ChannelRegTable.channelID] = channelID
            }
        }
        return true
    }

    suspend fun deleteChannel(userID:Int,botID:String,channelID:String) = dbQuery {
        val isAvailable = BotRegTable.select(BotRegTable.ownerID).where {
            (BotRegTable.id eq UUID.fromString(botID)) and (BotRegTable.ownerID eq userID)
        }.singleOrNull()
        if (isAvailable == null) return@dbQuery false
        ChannelRegTable.deleteWhere {
            (ChannelRegTable.botID eq UUID.fromString(botID)) and (ChannelRegTable.channelID eq channelID)
        }
    }

    private suspend fun youtubeVideoQuery(videoID: String): YtResponse.QueryData? {
        val res = client.get("https://www.googleapis.com/youtube/v3/videos") {
            url {
                parameters.append("part", "snippet,liveStreamingDetails")
                parameters.append("id", videoID)
                parameters.append("key", apiKey)
            }
        }

        val response = res.body<YtResponse.VideoListResponse>()
        val content = response.items.firstOrNull() ?: return null

        val videoDesc = content.snippet
        val thumb = videoDesc.thumbnails.maxres?.url
            ?: videoDesc.thumbnails.standard?.url
            ?: videoDesc.thumbnails.high?.url
            ?: videoDesc.thumbnails.default.url

        val startTime: String? = content.liveStreamingDetails?.actualStartTime
            ?: content.liveStreamingDetails?.scheduledStartTime
        val endTime: String? = content.liveStreamingDetails?.actualEndTime

        return YtResponse.QueryData(
            type = videoDesc.liveBroadcastContent,
            channelID = videoDesc.channelId,
            videoId = videoID,
            channelTitle = videoDesc.channelTitle,
            title = videoDesc.title,
            publishedAt = convertTimeEpochSec(videoDesc.publishedAt),
            startTime = startTime?.let { convertTimeEpochSec(it) },
            endTime = endTime?.let { convertTimeEpochSec(it) },
            thumbnail = thumb
        )
    }

    private fun buildWebhookPayload(
        queryData: YtResponse.QueryData,
        roleID: String,
        deliverState: DeliverState
    ): WebhookData {
        val vidUrl = "https://www.youtube.com/watch?v=${queryData.videoId}"
        val thumbUrl = queryData.thumbnail
        val fieldPlacer: MutableList<WebhookEmbed.Field> = mutableListOf()
        var msg = "<@&${roleID}> "

        if (roleID.contains("ignore:")){
            val fixedRole = roleID.removePrefix("ignore:")
            msg = fixedRole
        }


        val notifyType = deliverState.label

        when (deliverState) {
            DeliverState.UPCOMING -> {
                msg = msg.plus("配信枠が作成されました！")
                fieldPlacer.addLast(
                    WebhookEmbed.Field(
                        "開始時刻(予定)",
                        "<t:${queryData.startTime}:F> (<t:${queryData.startTime}:R>)"
                    )
                )
            }

            DeliverState.ON_AIR -> {
                msg = msg.plus("配信が開始されました！")
                fieldPlacer.addLast(WebhookEmbed.Field("開始時刻", "<t:${queryData.startTime}:F>"))
            }

            DeliverState.FINISHED -> {
                msg = "配信が終了しました。おつかれさまでした！"
                fieldPlacer.addLast(WebhookEmbed.Field("開始日時", "<t:${queryData.startTime}:F>"))
                fieldPlacer.addLast(WebhookEmbed.Field("終了時刻", "<t:${queryData.endTime}:T>"))
            }

            DeliverState.VIDEO -> {
                msg = msg.plus("動画が投稿されました！")
                fieldPlacer.addLast(WebhookEmbed.Field("投稿日時", "<t:${queryData.publishedAt}:F>"))
            }

            else -> {
                msg = "エラーが発生しました！"
                fieldPlacer.addLast(WebhookEmbed.Field("発生日時", "<t:${Instant.now().epochSecond}:F"))
            }
        }

        fieldPlacer.addLast(WebhookEmbed.Field("URL", vidUrl))


        return WebhookData(
            username = "${queryData.channelTitle} - $notifyType",
            embeds = listOf(
                WebhookEmbed.Embed(
                    title = "${queryData.title} [$notifyType]",
                    url = vidUrl,
                    color = deliverState.embedColor,
                    image = WebhookEmbed.Image(thumbUrl),
                    fields = fieldPlacer,
                    thumbnail = WebhookEmbed.Thumbnail(thumbUrl)
                )
            ),
            content = msg
        )
    }

    suspend fun notifyToBots(videoID: String): Boolean {
        val queryRes = youtubeVideoQuery(videoID) ?: return false
        val status = classifyStatusToNotify(queryRes)
        if (status.toDeliver) {
            val botHookList: Map<String, String> = dbQuery {
                (ChannelRegTable innerJoin BotRegTable)
                    .select(BotRegTable.wsUrl, BotRegTable.mentionRoleID)
                    .where { ChannelRegTable.channelID eq queryRes.channelID }
                    .associate { it[BotRegTable.wsUrl] to it[BotRegTable.mentionRoleID] }
            }
            for (query in botHookList) {
                client.post(query.key) {
                    contentType(ContentType.Application.Json)
                    setBody(buildWebhookPayload(queryRes, query.value, deliverState = status))
                }
            }

            return true
        }
        return false
    }
}