package net.kazugmx

import com.zaxxer.hikari.HikariConfig
import com.zaxxer.hikari.HikariDataSource
import io.ktor.client.HttpClient
import io.ktor.client.engine.cio.CIO
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation as clientContentNegotiation
import io.ktor.serialization.kotlinx.json.*
import io.ktor.serialization.kotlinx.xml.*
import io.ktor.server.application.*
import io.ktor.server.config.ApplicationConfig
import io.ktor.server.plugins.contentnegotiation.*
import io.ktor.server.plugins.openapi.openAPI
import io.ktor.server.routing.routing
import kotlinx.serialization.json.Json
import net.kazugmx.module.AuthService
import net.kazugmx.module.BotService
import net.kazugmx.module.configureAdministration
import net.kazugmx.schema.BotRegTable
import net.kazugmx.schema.ChannelRegTable
import net.kazugmx.schema.ChannelTable
import net.kazugmx.schema.MailConfig
import net.kazugmx.schema.OnAirTable
import net.kazugmx.schema.UserTable
import org.jetbrains.exposed.sql.Database
import org.jetbrains.exposed.sql.SchemaUtils
import org.jetbrains.exposed.sql.transactions.transaction
import org.koin.core.context.startKoin
import org.koin.dsl.module
import org.slf4j.LoggerFactory
import java.io.File
import java.security.SecureRandom


fun main(args: Array<String>) {
    io.ktor.server.netty.EngineMain.main(args)
}

fun ensureDBDirectory() {
    File("data/data_db").parentFile?.let {
        if (!it.exists()) it.mkdirs()
    }
}

private fun ApplicationConfig.readString(path: String, default: String): String =
    propertyOrNull(path)?.getString() ?: default

fun initTables(db: Database) {
    transaction {
        SchemaUtils.create(UserTable)
        SchemaUtils.create(ChannelTable)
        SchemaUtils.create(ChannelRegTable)
        SchemaUtils.create(BotRegTable)
        SchemaUtils.create(OnAirTable)
    }
}


fun Application.module() {

    val envJdbcUrl = environment.config.readString(
        "db.url",
        System.getenv("JDBC_URL") ?: "jdbc:sqlite:data/bot_data.db"
    )
    if (envJdbcUrl == "jdbc:sqlite:data/bot_data.db") ensureDBDirectory()
    val dbUser = environment.config.readString("db.user", System.getenv("DB_USER") ?: "")
    val dbPass = environment.config.readString("db.password", System.getenv("DB_PASSWORD") ?: "")

    val config = HikariConfig().apply {
        jdbcUrl = envJdbcUrl
        maximumPoolSize = 10
        when {
            envJdbcUrl.startsWith("jdbc:sqlite:") -> {
                driverClassName = "org.sqlite.JDBC"
            }

            envJdbcUrl.startsWith("jdbc:postgresql:") -> {
                require(dbUser != "invalidUser") { "DB_USER is required for PostgreSQL." }
                require(dbPass != "invalidPass") { "DB_PASSWORD is required for PostgreSQL." }
                driverClassName = "org.postgresql.Driver"
                username = dbUser
                password = dbPass
                transactionIsolation = "TRANSACTION_READ_COMMITTED"
                leakDetectionThreshold = 10_000
            }

            else -> {
                throw Exception("Invalid JDBC URL: $envJdbcUrl")
            }
        }
        connectionTestQuery = "SELECT 1"
        validate()
    }
    val apiKey = environment.config.readString(
        "youtube.apikey",
        System.getenv("APIKEY") ?: "test-api-key"
    )
    val rootOrigin = environment.config.readString(
        "youtube.callbackOrigin",
        System.getenv("CALLBACK_ORIGIN") ?: "localhost"
    )
    val uriMaster = environment.config.readString(
        "youtube.uri_master",
        System.getenv("URI_MASTER") ?: "https://example.com"
    )
    require(!(apiKey == "invalidKey" || rootOrigin == "invalidOrigin" || uriMaster == "invalidURIMaster")) {
        throw Exception("Environment arguments is invalid.")
    }


    val dataSource = HikariDataSource(config)

    val database = Database.connect(datasource = dataSource).also{
        initTables(it)
    }

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
        host = environment.config.readString("smtp.host", System.getenv("SMTP_HOST") ?: "none"),
        port = environment.config.readString("smtp.port", System.getenv("SMTP_PORT") ?: "-1").toInt(),
        user = environment.config.readString("smtp.user", System.getenv("SMTP_USER") ?: "none"),
        pass = environment.config.readString("smtp.password", System.getenv("SMTP_PASSWORD") ?: "none"),
        mailAddress = environment.config.readString("smtp.mailAddress", System.getenv("SMTP_MAIL_ADDRESS") ?: "none")
    )
    log.info(mailConfig.host)

    val module = module {
        single { database }
        single { SecureRandom.getInstanceStrong() }
        single {
            AuthService(
                db = get<Database>(),
                isMailActive = mailConfig.host != "none",
                mailConfig = mailConfig,
                logger = authLogger,
                newUser = environment.config.readString("appConfig.newUser", System.getenv("NEW_USER") ?: "false").toBoolean(),
                origin = rootOrigin,
                securePRNG = get()
            )
        }
        single {
            BotService(
                db = get<Database>(),
                logger = botLogger,
                apiKey = apiKey,
                origin = rootOrigin,
                client = get()
            )
        }
        single {
            HttpClient(CIO) {
                install(clientContentNegotiation) {
                    json(
                        Json {
                            ignoreUnknownKeys = true
                            isLenient = true
                        }
                    )
                }
            }
        }
    }

    startKoin {
        modules(module)
    }

    initAuthUnit()
    initBotUnit()
    initMiscUnit()
    configureAdministration(database)

    routing {
        openAPI(path = "/api/v1/docs", swaggerFile = "openapi/documentation.yaml")
    }
}