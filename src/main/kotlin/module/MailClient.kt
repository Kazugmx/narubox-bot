package net.kazugmx.module

import jakarta.mail.Session
import jakarta.mail.Message
import jakarta.mail.Authenticator
import jakarta.mail.internet.MimeMessage
import jakarta.mail.PasswordAuthentication
import jakarta.mail.Transport
import jakarta.mail.internet.InternetAddress
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import net.kazugmx.schema.MailConfig
import java.util.Properties

interface MailSender{
    suspend fun send(to: String, subject:String,body: String)
}

class SmtpMailSender(private val mailClient: MailClient): MailSender{
    override suspend fun send(to: String, subject: String, body: String) {
        mailClient.send(to, subject, body)
    }
}

object NoopMailSender: MailSender{
    override suspend fun send(to: String, subject: String, body: String) = Unit
}

class MailClient(private val config: MailConfig) {
    val props = Properties().apply {
        put("mail.smtp.host", config.host)
        put("mail.smtp.port", config.port.toString())
        put("mail.smtp.auth", true)
        put("mail.smtp.starttls.enable", "true")

        // TLS 固定
        put("mail.smtp.ssl.protocols", "TLSv1.2 TLSv1.3")
        put("mail.smtp.ssl.enable", "true")

        // Timeout
        put("mail.smtp.connectiontimeout", "10000")
        put("mail.smtp.timeout", "10000")
        put("mail.smtp.writetimeout", "10000")
    }

    suspend fun send(target: String, subjectCfg: String = "testSubject", text: String = "testText") {
        withContext(Dispatchers.IO) {

            val session = Session.getInstance(props, object : Authenticator() {
                override fun getPasswordAuthentication(): PasswordAuthentication {
                    return PasswordAuthentication(config.user, config.pass)
                }
            })

            val message = MimeMessage(session).apply {
                setFrom(InternetAddress(config.mailAddress))
                setRecipient(
                    Message.RecipientType.TO, InternetAddress(target)
                )
                subject = subjectCfg
                setText(text)
            }

            var lastException: Exception? = null
            for (i in 1..3) {
                try {
                    Transport.send(message)
                    return@withContext
                } catch (e: Exception) {
                    lastException = e
                    if (i < 3) kotlinx.coroutines.delay(2000L * i)
                }
            }
            throw lastException ?: Exception("Unknown error during mail transport")
        }
    }
}