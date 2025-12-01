package net.kazugmx.schema

import kotlinx.serialization.Serializable
import org.jetbrains.exposed.dao.id.IntIdTable
import org.jetbrains.exposed.dao.id.UUIDTable
import org.jetbrains.exposed.sql.Table
import org.jetbrains.exposed.sql.kotlin.datetime.CurrentDateTime
import org.jetbrains.exposed.sql.kotlin.datetime.datetime

@Serializable
data class BotRegisterReq(
    val botLabel: String,
    val wsUrl: String,
    val mentionRoleID: String
)

@Serializable
data class BotRegisterRes(
    val success: Boolean = false,
    val botID: String? = null
)

@Serializable
data class BotListRes(
    val botID: String,
    val label: String,
    val wsUrl: String,
    val mentionRoleID: String
)

@Serializable
data class SubscribeRequest(
    val channelID: String,
    val refresh: Boolean = false
)

object BotRegTable : UUIDTable("bot_table") {
    val ownerID = integer("owner_id").references(UserTable.id)
    val label = varchar(length = 50, name = "label")
    val wsUrl = varchar(length = 255, name = "ws_url")
    val mentionRoleID= varchar(length = 60, name = "mention_role_id")
}

object ChannelRegTable : IntIdTable("channel_bot_tags") {
    val botID = uuid("bot_id").references(BotRegTable.id)
    val channelID = varchar("channel_id", length = 60).references(ChannelTable.channelID)
}

object ChannelTable : Table("reg_channel") {
    val channelID = varchar("channel_id", length = 60)
    var lastUpdate = datetime("last_update").defaultExpression(CurrentDateTime)
}

object OnAirTable : Table("on_air") {
    val videoID = varchar("video_id", length = 20)
    val previousState = varchar("previous_state", length = 10)
}