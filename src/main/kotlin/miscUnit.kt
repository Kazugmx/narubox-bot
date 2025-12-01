package net.kazugmx

import io.ktor.http.HttpStatusCode
import io.ktor.server.application.Application
import io.ktor.server.response.respond
import io.ktor.server.routing.get
import io.ktor.server.routing.route
import io.ktor.server.routing.routing

fun Application.initMiscUnit() {
    routing{
        route("/misc"){
            get("naru-saikou"){
                call.respond(HttpStatusCode.Accepted,"ﾅﾙﾁｬﾝｶﾜｲｲﾔｯﾀｰ")
            }
            get("kemomimi"){
                call.respond(HttpStatusCode.Accepted,"けもみみもふもふ！！")
            }
        }
    }
}