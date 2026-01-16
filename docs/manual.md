# Manual
## Authentication
1. User registration
   POST to `/api/v1/auth/user` with JSON
    - Content-Type: Application/JSON
      Parameters
    - mail  (String): Value must match RFC 5322&5311
    - username (String)
    - password (String)
2. Mail verification (optional)
   info: This procedure is required if a mail server is configured with the application.
   Access to URL sent to your mail inbox.
   method : `GET`
   endpoint :  `/api/v1/auth/login`
3.  Login
    method : `GET`
    Content-Type: `Application/JSON`
    endpoint : `/api/v1/auth/login`
    parameters >
    - username (String)
    - password (String)

### Notes
- User registration requires a valid email address and password.
- Mail verification is only necessary if the application is configured with a mail server.
- Login requires valid credentials for authentication.
- The login endpoint returns a JWT token that is valid for 1 hour.
- The JWT token must be included in all requests to protected endpoints.
  - Header: "Authorization: Bearer {token}"

## Configuration

1. Bot registration
   Authentication is required with this action.
   method: `POST`
   Content-Type: `Application/JSON`
   endpoint: `/api/v1/bot/register`
   Parameters:
    - Bot configuration details (refer to BotRegisterReq schema)

2. Get bot list
   Authentication is required with this action.
   method: `GET`
   Content-Type: `Application/JSON`
   endpoint: `/api/v1/bot`
   Returns: List of bots registered by the authenticated user

3. Get channels for a bot
   Authentication is required with this action.
   method: `GET`
   Content-Type: `Application/JSON`
   endpoint: `/api/v1/bot/{botID}`
   Parameters:
    - botID (String): Bot identifier in URL path
      Returns: List of channels subscribed by the bot

4. Channel subscribe
   Authentication is required with this action.
   method: `POST`
   Content-Type: `Application/JSON`
   endpoint: `/api/v1/bot/{botID}`
   Parameters:
    - botID (String): Bot identifier in URL path
    - Channel subscription details (refer to SubscribeRequest schema)

5. Unregister bot
   Authentication is required with this action.
   method: `DELETE`
   Content-Type: `Application/JSON`
   endpoint: `/api/v1/bot/{botID}`
   Parameters:
    - botID (String): Bot identifier in URL path

6. Unsubscribe from channel
   Authentication is required with this action.
   method: `DELETE`
   Content-Type: `Application/JSON`
   endpoint: `/api/v1/bot/{botID}/channel/{channelID}`
   Parameters:
    - botID (String): Bot identifier in URL path
    - channelID (String): Channel identifier in URL path