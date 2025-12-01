#! /bin/bash
# use this script to start server instantly
# shellcheck disable=SC2034
APIKEY=$(cat .API_KEY) \
CALLBACK_ORIGIN=$(cat .CALLBACK_ORIGIN) \
JWT_SECRET=$(cat .JWT_SECRET) \
java -jar discordbot-narufc-all.jar
