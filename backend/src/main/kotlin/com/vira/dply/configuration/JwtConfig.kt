package com.vira.dply.configuration

import com.nimbusds.jose.JWSAlgorithm
import com.nimbusds.jose.jwk.OctetSequenceKey
import org.springframework.beans.factory.annotation.Value
import org.springframework.context.annotation.Configuration
import javax.crypto.SecretKey

@Configuration
class JwtConfig {
    @Value($$"${security.jwt.secret-key}")
    private val secretKey: String? = null

    @Value($$"${security.jwt.expiration-time}")
    private val jwtExpiration: Long = 0

    @Value($$"${security.jwt.algorithm}")
    private val algorithm: String? = null

    fun getSecretKey(): SecretKey {
        val key =
            OctetSequenceKey.Builder(secretKey!!.toByteArray())
                .algorithm(JWSAlgorithm(algorithm))
                .build()
        return key.toSecretKey()
    }

    fun getAlgorithm(): JWSAlgorithm? {
        return JWSAlgorithm(algorithm)
    }
}