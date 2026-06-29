# API Reference

This document summarizes the HTTP API exposed by narubox-bot.
The implementation is defined in the routing modules under the Kotlin source tree and the OpenAPI document at [src/main/resources/openapi/documentation.yaml](../src/main/resources/openapi/documentation.yaml).

## Base URL

- Local development: http://localhost:8080
- Production: the origin configured for the deployment environment

## Authentication

Most bot-management endpoints require a JWT bearer token.

Headers:

```http
Authorization: Bearer <jwt-token>
Content-Type: application/json
```

### Login flow

1. Register a user with POST /api/v1/auth/user.
2. If mail verification is enabled, open the verification URL sent by email.
3. Log in with POST /api/v1/auth/login to receive a JWT token.

---

## Auth endpoints

### POST /api/v1/auth/user

Create a new user account.

Request body:

```json
{
  "username": "example-user",
  "mail": "user@example.com",
  "password": "secret-password"
}
```

Responses:

- 201 Created: account created successfully
- 400/500: invalid input or registration failure

### GET /api/v1/auth/login

Verify an email-based registration token.

Query parameters:

- token: verification token sent by mail

Responses:

- 200 OK: verification succeeded
- 400 Bad Request: missing token
- 403 Forbidden: invalid or expired token

### POST /api/v1/auth/login

Authenticate an existing user and return a JWT.

Request body:

```json
{
  "username": "example-user",
  "password": "secret-password"
}
```

Responses:

- 200 OK: returns a token object

```json
{
  "token": "<jwt-token>"
}
```

- 401 Unauthorized: invalid credentials
- 403 Forbidden: email verification is still pending

### GET /api/v1/auth/user

Get the authenticated user's profile information.

Authentication: required

Responses:

- 200 OK: returns the user profile

```json
{
  "mail": "user@example.com",
  "username": "example-user",
  "createdAt": "2026-06-29T12:00:00",
  "lastAccess": null
}
```

---

## Bot management endpoints

### GET /api/v1/bot

List bots owned by the authenticated user.

Authentication: required

Responses:

- 200 OK: list of bot summaries

```json
[
  {
    "botID": "<uuid>",
    "label": "My Discord Bot",
    "mentionRoleID": "123456789"
  }
]
```

### POST /api/v1/bot/register

Register a new bot for the authenticated user.

Authentication: required

Request body:

```json
{
  "botLabel": "My Discord Bot",
  "wsUrl": "https://example.com/webhook",
  "mentionRoleID": "123456789"
}
```

Responses:

- 202 Accepted: registration succeeded
- 400 Bad Request: registration failed

### GET /api/v1/bot/{botID}

Get a specific bot and the channels subscribed to it.

Authentication: required

Path parameters:

- botID: bot identifier

Responses:

- 200 OK: bot details and subscribed channels

```json
{
  "botInfo": {
    "botID": "<uuid>",
    "label": "My Discord Bot",
    "mentionRoleID": "123456789"
  },
  "channels": [
    "UC1234567890"
  ]
}
```

### POST /api/v1/bot/{botID}

Subscribe a YouTube channel to a bot.

Authentication: required

Path parameters:

- botID: bot identifier

Request body:

```json
{
  "channelID": "UC1234567890",
  "refresh": false
}
```

Responses:

- 202 Accepted: subscription succeeded
- 400 Bad Request: subscription failed

### DELETE /api/v1/bot/{botID}

Unregister a bot.

Authentication: required

Path parameters:

- botID: bot identifier

Responses:

- 202 Accepted: bot removed
- 400 Bad Request: invalid request

### DELETE /api/v1/bot/{botID}/channel/{channelID}

Unsubscribe a channel from a bot.

Authentication: required

Path parameters:

- botID: bot identifier
- channelID: subscribed channel identifier

Responses:

- 202 Accepted: unsubscribe succeeded
- 400 Bad Request: invalid request

### GET /api/v1/bot/pubsub/{endpointID}

PubSubHubbub challenge endpoint used during subscription verification.

Query parameters:

- hub.challenge: challenge value returned by the hub

Responses:

- 200 OK: returns the challenge value as plain text
- 400 Bad Request: missing challenge parameter

### POST /api/v1/bot/pubsub/{endpointID}

Receive PubSubHubbub XML notifications from YouTube.

Path parameters:

- endpointID: endpoint identifier generated for the channel subscription

Request body:

- XML feed body from YouTube PubSubHubbub

Responses:

- 200 OK: notification accepted

---

## Misc endpoints

### GET /misc/naru-saikou

Returns a fun response string.

Responses:

- 202 Accepted

### GET /misc/kemomimi

Returns a fun response string.

Responses:

- 202 Accepted

---

## Request/response schemas

### UserCreateReq

```json
{
  "username": "string",
  "mail": "user@example.com",
  "password": "string"
}
```

### LoginReq

```json
{
  "username": "string",
  "password": "string"
}
```

### BotRegisterReq

```json
{
  "botLabel": "string",
  "wsUrl": "string",
  "mentionRoleID": "string"
}
```

### SubscribeRequest

```json
{
  "channelID": "string",
  "refresh": false
}
```

### SelfRes

```json
{
  "mail": "string",
  "username": "string",
  "createdAt": "datetime",
  "lastAccess": "datetime|null"
}
```

---

## Notes

- The server uses JWT-based authentication for protected endpoints.
- The PubSubHubbub callback path is generated automatically for each subscribed channel.
- The OpenAPI document is available at [src/main/resources/openapi/documentation.yaml](../src/main/resources/openapi/documentation.yaml) and can be used with Swagger UI or Redoc.
