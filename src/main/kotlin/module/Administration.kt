package net.kazugmx.module

import io.github.flaxoos.ktor.server.plugins.taskscheduling.TaskScheduling
import io.github.flaxoos.ktor.server.plugins.taskscheduling.managers.lock.database.DefaultTaskLockTable
import io.github.flaxoos.ktor.server.plugins.taskscheduling.managers.lock.database.jdbc
import io.ktor.server.application.*
import org.jetbrains.exposed.sql.Database
import org.jetbrains.exposed.sql.SchemaUtils
import org.jetbrains.exposed.sql.transactions.transaction
import org.koin.ktor.ext.inject
import java.time.ZoneId
import java.time.format.DateTimeFormatter
import kotlin.time.Instant
import kotlin.time.toJavaInstant

fun Application.configureAdministration(db: Database) {
    val botService:BotService by inject()
    install(TaskScheduling) {
        jdbc {
            database = db.also {
                transaction { SchemaUtils.create(DefaultTaskLockTable) }
            }
        }

        task {
            name = "RefreshChannelCallback"
            task = { taskExecutionTime ->
                val jInstant = Instant.fromEpochMilliseconds(taskExecutionTime.unixMillisLong)
                botService.refreshChannel()

                log.info(
                    "My task is running persec: ${
                        jInstant.toJavaInstant().atZone(ZoneId.systemDefault()).format(
                            DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss.SSS")
                        )
                    }"
                )
            }
            kronSchedule = {
                seconds { at(0) }
                minutes { at(0) }
                hours { at(0) }
                dayOfWeek { at(0)}
            }

            concurrency = 1
        }
    }
}