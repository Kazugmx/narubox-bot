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

class MailClient(private val config: MailConfig) {
    suspend fun send(target: String) {
        withContext(Dispatchers.IO) {
            val props = Properties().apply {
                put("mail.smtp.host", config.host)
                put("mail.smtp.port", config.port)
                put("mail.smtp.auth", true)
                put("mail.smtp.starttls.enable", "true")
            }
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
                subject = "Test Mail"
                setText("Hello World!")
            }

            Transport.send(message)
        }
    }
}