FROM eclipse-temurin:21-jre-jammy AS builder

WORKDIR /app
COPY build/libs/discordbot-narufc-all.jar app.jar
COPY src/main/resources/application.yaml /app/application.yaml
COPY src/main/resources/logback.xml /app/logback.xml

FROM gcr.io/distroless/java21
WORKDIR /app
COPY --from=builder /app /app

EXPOSE 8080
ENV APIKEY=invalidKey
ENV CALLBACK_ORIGIN=invalidOrigin

CMD ["app.jar"]