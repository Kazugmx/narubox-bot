val exposed_version: String by project
val h2_version: String by project
val koin_version: String by project
val kotlin_version: String by project
val logback_version: String by project
val postgres_version: String by project
val bcrypt_version: String by project

plugins {
    kotlin("jvm") version "2.2.20"
    id("io.ktor.plugin") version "3.3.2"
    id("org.jetbrains.kotlin.plugin.serialization") version "2.2.20"
    id("com.gradleup.shadow") version "9.2.2"
    application
}

kotlin{
    jvmToolchain(21)
}

group = "net.kazugmx"
version = "1.1.0-PROD"

application {
    mainClass = "net.kazugmx.ApplicationKt"
}

dependencies {
    implementation("net.dv8tion:JDA:6.1.1")
    implementation("at.favre.lib:bcrypt:${bcrypt_version}")

    implementation("org.openfolder:kotlin-asyncapi-ktor:3.1.3")
    implementation("io.ktor:ktor-server-cors")
    implementation("io.ktor:ktor-server-default-headers")
    implementation("io.ktor:ktor-server-forwarded-header")
    implementation("io.ktor:ktor-server-core")
    implementation("io.ktor:ktor-server-openapi")
    implementation("com.ucasoft.ktor:ktor-simple-cache:0.55.3")
    implementation("com.ucasoft.ktor:ktor-simple-redis-cache:0.55.3")
    implementation("io.ktor:ktor-server-auth")
    implementation("io.ktor:ktor-server-auth-jwt")
    implementation("io.ktor:ktor-server-content-negotiation")
    implementation("io.ktor:ktor-serialization-kotlinx-json")
    implementation("io.ktor:ktor-server-sessions")
    implementation("io.ktor:ktor-server-request-validation")
    implementation("io.ktor:ktor-server-call-logging")
    implementation("io.ktor:ktor-server-metrics")
    implementation("io.ktor:ktor-client-core")
    implementation("io.ktor:ktor-client-cio")
    implementation("org.jetbrains.exposed:exposed-core:$exposed_version")
    implementation("org.jetbrains.exposed:exposed-jdbc:$exposed_version")
    implementation("com.h2database:h2:$h2_version")
    implementation("com.zaxxer:HikariCP:7.0.2")
    implementation("org.xerial:sqlite-jdbc:3.50.3.0")
    implementation("org.postgresql:postgresql:$postgres_version")
    implementation("io.ktor:ktor-server-di")
    implementation("io.insert-koin:koin-ktor:$koin_version")
    implementation("io.insert-koin:koin-logger-slf4j:$koin_version")
    implementation("io.github.damirdenis-tudor:ktor-server-rabbitmq:1.3.6")
    implementation("io.github.flaxoos:ktor-server-task-scheduling-core:2.2.1")
    implementation("io.github.flaxoos:ktor-server-task-scheduling-redis:2.2.1")
    implementation("io.github.flaxoos:ktor-server-task-scheduling-mongodb:2.2.1")
    implementation("io.github.flaxoos:ktor-server-task-scheduling-jdbc:2.2.1")
    implementation("io.ktor:ktor-server-netty")
    implementation("ch.qos.logback:logback-classic:$logback_version")
    implementation("io.ktor:ktor-server-config-yaml")
    implementation("io.ktor:ktor-client-content-negotiation:3.3.2")
    implementation("io.ktor:ktor-serialization-kotlinx-xml:3.3.2")
    testImplementation("io.ktor:ktor-server-test-host")
    testImplementation("org.jetbrains.kotlin:kotlin-test-junit:$kotlin_version")
}

tasks.shadowJar {
    archiveBaseName.set("kazunyan-youtube-pubsub-bot")
    archiveClassifier.set("")
    archiveVersion.set(version.toString())
}