## auth
- token w/ JWT
- /api/v1/auth
    - user @Requires auth
        - GET
        - POST
    - login
        - GET
        For mail verification.
        Query Param
        1. token (string)
        - POST
        Auth endpoint.
        ContentType: Application/JSON
        Payload:
        { "username": string, "password": string}
- /api/v1/bot
    - pubsub/{endpointID}
        - GET
        endpoint for PubSub challenge
        Query Param>
        1. hub.challenge (string)
        - POST
        receive content data from PubSub server
    - GET
    @Requires auth
    retrieve owned bots by user
    - register(POST)
    