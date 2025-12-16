# --- build ---
FROM eclipse-temurin:21-jdk AS builder
WORKDIR /build
COPY . .
RUN ./gradlew clean buildFatJar --no-daemon

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
ENV JAVA_OPTS="-Xms128m -Xmx512m -XX:MaxRAMPercentage=75.0 -XX:InitialRAMPercentage=25.0"

CMD ["java","-jar","app.jar"]