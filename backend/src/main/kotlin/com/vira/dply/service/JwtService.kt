package com.vira.dply.service

import io.jsonwebtoken.Claims;
import io.jsonwebtoken.Jwts;
import io.jsonwebtoken.SignatureAlgorithm;
import io.jsonwebtoken.io.Decoders;
import io.jsonwebtoken.security.Keys;
import jakarta.servlet.http.Cookie;
import jakarta.servlet.http.HttpServletRequest;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.ResponseCookie;
import org.springframework.security.core.userdetails.UserDetails;
import org.springframework.stereotype.Service;
import org.springframework.web.util.WebUtils;

import java.security.Key;
import java.util.Date;
import java.util.HashMap;
import java.util.function.Function;

@Service
class JwtService {

    @Value("\${application.security.jwt.secret-key}")
    private val secretKey: String? = null

    @Value($$"${application.security.jwt.expiration}")
    private val jwtExpiration: Long = 0

    @Value("\${application.security.jwt.cookie-name}")
    private val jwtCookieName: String? = null

    public fun extractUserName(token: String?): String {
        return extractClaim<String?>(token) { it?.getSubject() } ?: ""
    }

    public fun generateToken(userDetails: UserDetails): String {
        return generateToken(HashMap<String?, Any?>(), userDetails)
    }

    public fun isTokenValid(token: String?, userDetails: UserDetails): Boolean {
        val userName = extractUserName(token)
        return (userName == userDetails.username) && !isTokenExpired(token)
    }

    private fun isTokenExpired(token: String?): Boolean {
        val expiration = extractExpiration(token)
        return expiration == null || expiration.before(Date())
    }

    private fun extractExpiration(token: String?): Date? {
        return extractClaim<Date?>(token) { it?.getExpiration() }
    }

    private fun generateToken(extraClaims: MutableMap<String?, Any?>?, userDetails: UserDetails): String {
        return buildToken(extraClaims, userDetails, jwtExpiration)
    }

    public fun generateJwtCookie(jwt: String?): ResponseCookie {
        return ResponseCookie.from(jwtCookieName!!, jwt)
            .path("/")
            .maxAge((24 * 60 * 60).toLong()) // 24 hours
            .httpOnly(true)
            .secure(true)
            .sameSite("Strict")
            .build()
    }

    public fun getJwtFromCookies(request: HttpServletRequest): String? {
        val cookie: Cookie? = WebUtils.getCookie(request, jwtCookieName!!)
        return if (cookie != null) {
            cookie.value
        } else {
            null
        }
    }

    public fun getCleanJwtCookie(): ResponseCookie {
        return ResponseCookie.from(jwtCookieName!!, "")
            .path("/")
            .httpOnly(true)
            .maxAge(0)
            .build()
    }

    private fun buildToken(
        extraClaims: MutableMap<String?, Any?>?,
        userDetails: UserDetails,
        expiration: Long
    ): String {
        return Jwts
            .builder()
            .setClaims(extraClaims)
            .setSubject(userDetails.username)
            .setIssuedAt(Date(System.currentTimeMillis()))
            .setExpiration(Date(System.currentTimeMillis() + expiration))
            .signWith(getSigningKey(), SignatureAlgorithm.HS256)
            .compact()
    }

    private fun <T> extractClaim(token: String?, claimsResolvers: Function<Claims?, T?>): T? {
        val claims: Claims = extractAllClaims(token)
        return claimsResolvers.apply(claims)
    }

    private fun extractAllClaims(token: String?): Claims {
        return Jwts.parserBuilder()
            .setSigningKey(getSigningKey())
            .build()
            .parseClaimsJws(token)
            .getBody()
    }

    private fun getSigningKey(): Key {
        val keyBytes: ByteArray? = Decoders.BASE64.decode(secretKey)
        return Keys.hmacShaKeyFor(keyBytes)
    }
}