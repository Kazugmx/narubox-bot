# narubox-bot

narubox-bot for Discord.
This bot uses [Google PubSubHubbub](https://pubsubhubbub.appspot.com/) and
the [YouTube Data API](https://developers.google.com/youtube/v3) to send
real-time notifications to Discord about channel livestreams and video updates.

## Quickstart

1. run `./gradlew buildFatJar`
2. write config on `.API_KEY`, `.CALLBACK_ORIGIN`, `.JWT_SECRET`
3. run `java -jar` with jar files which you built
4. expose listen-port with proxy e.g. Nginx, Cloudflare Tunnel
5. enjoy!

manual : [/docs/manual.md](docs/manual.md)

## Features

- Realtime Notification for YouTube Live Stream & video uploads
- Discord Webhook Integration

| Name                                                                   | Description                                                                        |
|------------------------------------------------------------------------|------------------------------------------------------------------------------------|
| [Routing](https://start.ktor.io/p/routing)                             | Provides a structured routing DSL                                                  |
| [Authentication](https://start.ktor.io/p/auth)                         | Provides extension point for handling the Authorization header                     |
| [Authentication JWT](https://start.ktor.io/p/auth-jwt)                 | Handles JSON Web Token (JWT) bearer authentication scheme                          |
| [Content Negotiation](https://start.ktor.io/p/content-negotiation)     | Provides automatic content conversion according to Content-Type and Accept headers |
| [kotlinx.serialization](https://start.ktor.io/p/kotlinx-serialization) | Handles JSON serialization using kotlinx.serialization library                     |
| [Sessions](https://start.ktor.io/p/ktor-sessions)                      | Adds support for persistent sessions through cookies or headers                    |
| [Request Validation](https://start.ktor.io/p/request-validation)       | Adds validation for incoming requests                                              |
| [Exposed](https://start.ktor.io/p/exposed)                             | Adds Exposed database to your application                                          |

## Tech Stack

| Name          | Technology                                 |
|---------------|--------------------------------------------|
| Language      | Kotlin                                     |
| Web framework | [Ktor (server & client)](https://ktor.io/) |
| Database      | SQLite + Exposed ORM + HikariCP            |
| Auth          | JWT (Ktor authentication) + BCrypt         |
| logging       | Logback,SLF4J                              |
| API           | YouTube Data API v3                        |

## Building & Running

To build or run the project, use one of the following tasks:

| Task                                    | Description                                                          |
|-----------------------------------------|----------------------------------------------------------------------|
| `./gradlew test`                        | Run the tests                                                        |
| `./gradlew build`                       | Build everything                                                     |
| `./gradlew buildFatJar`                 | Build an executable JAR of the server with all dependencies included |
| `./gradlew buildImage`                  | Build the docker image to use with the fat JAR                       |
| `./gradlew publishImageToLocalRegistry` | Publish the docker image locally                                     |
| `./gradlew run`                         | Run the server                                                       |
| `./gradlew runDocker`                   | Run using the local docker image                                     |

If the server starts successfully, you'll see the following output:

```
2024-12-04 14:32:45.584 [main] INFO  Application - Application started in 0.303 seconds.
2024-12-04 14:32:45.682 [main] INFO  Application - Responding at http://0.0.0.0:8080
```

