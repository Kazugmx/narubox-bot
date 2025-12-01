package net.kazugmx.schema

import kotlinx.serialization.Serializable
import net.kazugmx.schema.WebhookEmbed.Embed
import nl.adaptivity.xmlutil.serialization.XmlElement
import nl.adaptivity.xmlutil.serialization.XmlSerialName

@Serializable
@XmlSerialName("feed", namespace = "http://www.w3.org/2005/Atom")
data class Feed(
    @XmlSerialName("title", namespace = "http://www.w3.org/2005/Atom")
    @XmlElement(true)
    val title: String? = null,

    @XmlSerialName("updated", namespace = "http://www.w3.org/2005/Atom")
    @XmlElement(true)
    val updated: String? = null,

    @XmlSerialName("link", namespace = "http://www.w3.org/2005/Atom")
    @XmlElement(true)
    val link: List<Link>? = null,

    @XmlSerialName("entry", namespace = "http://www.w3.org/2005/Atom")
    @XmlElement(true)
    val entry: List<Entry>? = null
)

@Serializable
@Suppress("HttpUrlsUsage")
@XmlSerialName("entry", namespace = "http://www.w3.org/2005/Atom")
data class Entry(
    @XmlSerialName("id", namespace = "http://www.w3.org/2005/Atom")
    @XmlElement(true)
    val id: String? = null,

    @XmlSerialName("title", namespace = "http://www.w3.org/2005/Atom")
    @XmlElement(true)
    val title: String? = null,

    @XmlSerialName("published", namespace = "http://www.w3.org/2005/Atom")
    @XmlElement(true)
    val publishedAt: String? = null,

    @XmlSerialName("updated", namespace = "http://www.w3.org/2005/Atom")
    @XmlElement(true)
    val updated: String? = null,

    @XmlSerialName("videoId", namespace = "http://www.youtube.com/xml/schemas/2015")
    @XmlElement(true)
    val videoID: String? = null,

    @XmlSerialName("channelId", namespace = "http://www.youtube.com/xml/schemas/2015")
    @XmlElement(true)
    val channelID: String? = null,

    @XmlSerialName("link", namespace = "http://www.w3.org/2005/Atom")
    @XmlElement(true)
    val link: List<Link>? = null,

    @XmlSerialName("author", namespace = "http://www.w3.org/2005/Atom")
    @XmlElement(true)
    val author: List<Author>? = null
)

@Serializable
@XmlSerialName("link", namespace = "http://www.w3.org/2005/Atom")
data class Link(
    @XmlSerialName("rel", namespace = "", prefix = "")
    val rel: String? = null,

    @XmlSerialName("href", namespace = "", prefix = "")
    val href: String? = null
)

@Serializable
@XmlSerialName("author", namespace = "http://www.w3.org/2005/Atom")
data class Author(
    @XmlSerialName("name", namespace = "http://www.w3.org/2005/Atom")
    @XmlElement(true)
    val name: String? = null,

    @XmlSerialName("uri", namespace = "http://www.w3.org/2005/Atom")
    @XmlElement(true)
    val uri: String? = null
)

// serialize youtubeAPI response


object YtResponse {
    @Serializable
    data class VideoListResponse(
        val items: List<Content> = emptyList(),
    )

    @Serializable
    data class Content(
        val id: String,
        val snippet: Snippet,
        val liveStreamingDetails: LiveStreamingDetails? = null
    )

    @Serializable
    data class Snippet(
        val channelId: String,
        val title: String,
        val channelTitle: String,
        val liveBroadcastContent: String,
        val thumbnails: Thumbnails,
        val publishedAt: String
    )

    @Serializable
    data class Thumbnails(
        val default: Thumbnail,
        val standard: Thumbnail? = null,
        val maxres: Thumbnail? = null,
        val high: Thumbnail? = null,
        val medium: Thumbnail? = null
    )

    @Serializable
    data class Thumbnail(
        val url: String
    )

    @Serializable
    data class LiveStreamingDetails(
        val actualStartTime: String? = null,
        val scheduledStartTime: String? = null,
        val actualEndTime: String? = null
    )

    data class QueryData(
        val type: String,
        val channelID: String,
        val videoId: String,
        val channelTitle: String,
        val title: String,
        val publishedAt: Long,
        val startTime: Long? = null,
        val endTime: Long? = null,
        val thumbnail: String
    )
}

@Serializable
data class WebhookData(
    val username: String,
    val embeds: List<Embed>,
    val attachments: List<String> = emptyList(),
    val content: String? = null
)

object WebhookEmbed {
    @Serializable
    data class Embed(
        val title: String,
        val url: String,
        val color: Int,
        val image: Image?,
        val fields: List<Field>,
        val thumbnail: Thumbnail?
    )

    @Serializable
    data class Image(val url: String)

    @Serializable
    data class Field(
        val name: String,
        val value: String
    )

    @Serializable
    data class Thumbnail(val url: String)
}

enum class DeliverState(val text: String, val label: String, val embedColor: Int = 0xFFFFFF, val toDeliver: Boolean = true) {
    UPCOMING("upcoming", "配信予告",embedColor = 0x050E3C),
    ON_AIR("live", "配信開始",embedColor = 0xFF3838),
    FINISHED("none", "配信終了", embedColor = 0x1B3C53),
    VIDEO("none", "動画投稿", embedColor = 0x6DC3BB),
    ERROR("invalid", "エラー", toDeliver = false)
}