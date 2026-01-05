# --- build ---
FROM eclipse-temurin:21-jdk AS builder
WORKDIR /build


COPY gradlew settings.gradle* build.gradle* gradle.properties* /build/
COPY gradle /build/gradle

RUN --mount=type=cache,target=/root/.gradle \
    ./gradlew --no-daemon -q help

COPY . /build/

RUN --mount=type=cache,target=/root/.gradle \
    ./gradlew buildFatJar --no-daemon

# --- runtime ---
FROM gcr.io/distroless/java21
WORKDIR /app
COPY --from=builder /build/build/libs/*.jar app.jar

EXPOSE 8080

ENV SERVER_PORT=8080
ENV APIKEY=invalidKey
ENV CALLBACK_ORIGIN=invalidOrigin
ENV JWT_SECRET=invalidSecret
ENV URI_MASTER=invalidURIMaster
ENV JAVA_TOOL_OPTIONS="-Xms128m -Xmx512m -XX:MaxRAMPercentage=75.0 -XX:InitialRAMPercentage=25.0"

ENTRYPOINT ["java","-jar","app.jar"]
