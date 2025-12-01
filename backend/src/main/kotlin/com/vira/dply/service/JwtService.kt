package com.vira.dply.service

import com.nimbusds.jose.JOSEException
import com.nimbusds.jose.JOSEObjectType
import com.nimbusds.jose.JWSHeader
import com.nimbusds.jose.crypto.MACSigner
import com.nimbusds.jwt.JWTClaimsSet
import com.nimbusds.jwt.SignedJWT
import com.vira.dply.configuration.JwtConfig
import com.vira.dply.model.User
import org.springframework.security.core.Authentication
import org.springframework.security.core.GrantedAuthority
import org.springframework.stereotype.Service
import java.time.Instant
import java.time.temporal.ChronoUnit
import java.util.Date


@Service
class JwtService(val jwtConfig: JwtConfig) {

    fun generateToken(authentication: Authentication): String {
        // header + payload/claims + signature
        val header = JWSHeader.Builder(jwtConfig.getAlgorithm())
            .type(JOSEObjectType.JWT)
            .build()
        val now: Instant = Instant.now()
        val roles =
            authentication.authorities.stream()
                .map { obj: GrantedAuthority? -> obj!!.authority }
                .toList()
        val builder = JWTClaimsSet.Builder()
            .issuer("VIRA-DPLY")
            .issueTime(Date.from(now))
            .expirationTime(Date.from(now.plus(1, ChronoUnit.HOURS)))
        builder.claim("roles", roles)
        val user: User = authentication.principal as User
        builder.claim("name", user.fullName)
        builder.claim("email", user.email)
        builder.claim("id", user.id)
        val claims = builder.build()

        val key = jwtConfig.getSecretKey()

        val jwt = SignedJWT(header, claims)

        try {
            val signer = MACSigner(key)
            jwt.sign(signer)
        } catch (e: JOSEException) {
            throw RuntimeException("Error generating JWT", e)
        }
        return jwt.serialize()
    }
}