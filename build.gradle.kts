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
    id("org.jetbrains.kotlin.plugin.serialization") version "2.3.0"
    id("com.gradleup.shadow") version "9.2.2"
    application
}

kotlin{
    jvmToolchain(21)
}

group = "net.kazugmx.narubox-bot"
version = "1.3.8"

application {
    mainClass = "net.kazugmx.ApplicationKt"
}

dependencies {
    implementation(platform("org.jetbrains.exposed:exposed-bom:$exposed_version"))
    implementation("net.dv8tion:JDA:6.1.1")
    implementation("at.favre.lib:bcrypt:${bcrypt_version}")

    implementation("org.openfolder:kotlin-asyncapi-ktor:3.1.3")
    implementation("io.ktor:ktor-server-cors")
    implementation("io.ktor:ktor-server-default-headers")
    implementation("io.ktor:ktor-server-forwarded-header")
    implementation("io.ktor:ktor-server-core")
    implementation("io.ktor:ktor-server-openapi")
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
    implementation("org.jetbrains.exposed:exposed-core")
    implementation("org.jetbrains.exposed:exposed-jdbc")
    implementation("org.jetbrains.exposed:exposed-dao")
    implementation("org.jetbrains.exposed:exposed-kotlin-datetime")
    implementation("com.h2database:h2:$h2_version")
    implementation("com.zaxxer:HikariCP:7.0.2")
    implementation("org.postgresql:postgresql:$postgres_version")
    implementation("io.ktor:ktor-server-netty")
    implementation("ch.qos.logback:logback-classic:$logback_version")
    implementation("io.ktor:ktor-server-config-yaml")
    implementation("io.ktor:ktor-client-content-negotiation:3.3.2")
    implementation("io.ktor:ktor-serialization-kotlinx-xml:3.3.2")
    api("org.jetbrains.kotlinx:kotlinx-datetime:0.6.2")
    implementation("com.sun.mail:jakarta.mail:2.0.2")
    implementation("io.github.flaxoos:ktor-server-task-scheduling-core:2.2.1")
    implementation("io.github.flaxoos:ktor-server-task-scheduling-jdbc:2.2.1")
    implementation("io.insert-koin:koin-ktor:4.1.2-Beta1")
    implementation("io.insert-koin:koin-logger-slf4j:4.1.2-Beta1")
    testImplementation("io.ktor:ktor-server-test-host")
    testImplementation("org.jetbrains.kotlin:kotlin-test-junit:$kotlin_version")


    //JDBC Runtimes
    runtimeOnly("org.postgresql:postgresql:42.7.7")
    runtimeOnly("org.xerial:sqlite-jdbc:3.50.3.0")
}
repositories {
    maven {
        url = uri("https://repo1.maven.org/maven2")
        name = "MavenCentral"
    }
    maven {
        url = uri("https://packages.confluent.io/maven")
        name = "confluent"
    }
}

tasks.shadowJar {
    archiveBaseName.set("kazunyan-youtube-pubsub-bot")
    archiveClassifier.set("server")
    archiveVersion.set(version.toString())
}