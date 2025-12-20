package net.kazugmx

import com.zaxxer.hikari.HikariConfig
import com.zaxxer.hikari.HikariDataSource
import io.ktor.serialization.kotlinx.json.*
import io.ktor.serialization.kotlinx.xml.*
import io.ktor.server.application.*
import io.ktor.server.plugins.contentnegotiation.*
import io.ktor.server.plugins.openapi.openAPI
import io.ktor.server.routing.routing
import kotlinx.serialization.json.Json
import net.kazugmx.module.AuthService
import net.kazugmx.module.BotService
import net.kazugmx.schema.MailConfig
import org.jetbrains.exposed.sql.Database
import org.slf4j.LoggerFactory
import java.io.File


fun main(args: Array<String>) {
    io.ktor.server.netty.EngineMain.main(args)
}

fun ensureDBDirectory() {
    File("data/data_db").parentFile?.let {
        if (!it.exists()) it.mkdirs()
    }
}

enum class DBType {
    SQLITE, POSTGRESQL, INVALID
}

fun Application.module() {

    val envJdbcUrl = environment.config.property("db.url").getString()
    if (envJdbcUrl == "jdbc:sqlite:data/bot_data.db") ensureDBDirectory()
    val dbUser = environment.config.property("db.user").getString()
    val dbPass = environment.config.property("db.password").getString()
    val dbType = when {
        envJdbcUrl.startsWith("jdbc:sqlite:") -> DBType.SQLITE
        envJdbcUrl.startsWith("jdbc:postgresql:") -> DBType.POSTGRESQL
        else -> DBType.INVALID
    }

    val config = HikariConfig().apply {
        jdbcUrl = envJdbcUrl
        maximumPoolSize = 10
        when (dbType) {
            DBType.SQLITE -> {
                driverClassName = "org.sqlite.JDBC"
            }

            DBType.POSTGRESQL -> {
                require(dbUser != "invalidUser") { "DB_USER is required for PostgreSQL." }
                require(dbPass != "invalidPass") { "DB_PASSWORD is required for PostgreSQL." }
                driverClassName = "org.postgresql.Driver"
                username = dbUser
                password = dbPass
                transactionIsolation = "TRANSACTION_READ_COMMITTED"
                leakDetectionThreshold = 10_000
            }

            else -> {
                throw Exception("Invalid JDBC URL: $envJdbcUrl / DBType: ${dbType.name}")
            }
        }
        connectionTestQuery = "SELECT 1"
        validate()
    }
    val apiKey = environment.config.property("youtube.apikey").getString()
    val rootOrigin = environment.config.property("youtube.callbackOrigin").getString()
    val uriMaster = environment.config.property("youtube.uri_master").getString()
    require(!(apiKey == "invalidKey" || rootOrigin == "invalidOrigin" || uriMaster == "invalidURIMaster")) {
        throw Exception("Environment arguments is invalid.")
    }


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

    val mailConfig = MailConfig(
        host = environment.config.property("smtp.host").getString(),
        port = environment.config.property("smtp.port").getString().toInt(),
        user = environment.config.property("smtp.user").getString(),
        pass = environment.config.property("smtp.pass").getString(),
        mailAddress = environment.config.property("smtp.mailAddress").getString()
    )

    val authSvc = AuthService(
        db = database,
        isMailActive = mailConfig.host != "none",
        mailConfig = mailConfig,
        logger = authLogger,
        newUser = environment.config.property("appConfig.newUser").getString().lowercase() != "false",
    )
    val botSvc = BotService(db = database, logger = botLogger, apiKey = apiKey, origin = rootOrigin)

    initAuthUnit(authSvc)
    initBotUnit(botSvc)
    initMiscUnit()

    routing {
        openAPI(path = "/api/v1/docs", swaggerFile = "openapi/documentation.yaml")
    }
}
