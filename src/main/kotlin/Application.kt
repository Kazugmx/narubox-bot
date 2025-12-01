package net.kazugmx

import com.zaxxer.hikari.HikariConfig
import com.zaxxer.hikari.HikariDataSource
import io.ktor.serialization.kotlinx.json.json
import io.ktor.serialization.kotlinx.xml.xml
import io.ktor.server.application.*
import io.ktor.server.plugins.contentnegotiation.ContentNegotiation
import kotlinx.serialization.json.Json
import org.jetbrains.exposed.sql.Database
import net.kazugmx.module.AuthService
import net.kazugmx.module.BotService
import org.slf4j.LoggerFactory


fun main(args: Array<String>) {
    io.ktor.server.netty.EngineMain.main(args)
}

fun Application.module() {
    val config =
        HikariConfig().apply {
            jdbcUrl = environment.config.property("db.url").getString()
            driverClassName = "org.sqlite.JDBC"
            maximumPoolSize = 10
            isAutoCommit = true
            transactionIsolation = "TRANSACTION_SERIALIZABLE"
            validate()
        }
    val apiKey = environment.config.property("youtube.apikey").getString()
    val rootOrigin = environment.config.property("youtube.callbackOrigin").getString()
    if (apiKey == "invalidKey") throw Exception("API Key is invalid.")
    if (rootOrigin == "invalidOrigin") throw Exception("Root Origin is invalid.")
    val dataSource = HikariDataSource(config)

    val database = Database.connect(datasource = dataSource)

    install(ContentNegotiation) {
        json(
            Json {
                ignoreUnknownKeys = true
                prettyPrint = true
                isLenient = true
            }
        )
        xml()
    }

    val authLogger = LoggerFactory.getLogger("AuthService")
    val botLogger = LoggerFactory.getLogger("BotService")

    val authSvc = AuthService(db = database, logger=authLogger)
    val botSvc = BotService(db = database, logger=botLogger, apiKey = apiKey,origin = rootOrigin)

    initAuthUnit(authSvc)
    initBotUnit(botSvc)
    initMiscUnit()
}
