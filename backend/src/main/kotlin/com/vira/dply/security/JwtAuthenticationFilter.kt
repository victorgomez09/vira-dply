package com.vira.dply.security

import com.vira.dply.service.JwtService
import com.vira.dply.service.UserService
import io.micrometer.common.util.StringUtils
import jakarta.servlet.FilterChain
import jakarta.servlet.http.HttpServletRequest
import jakarta.servlet.http.HttpServletResponse
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken
import org.springframework.security.core.context.SecurityContextHolder
import org.springframework.security.core.userdetails.UserDetails
import org.springframework.security.web.authentication.WebAuthenticationDetailsSource
import org.springframework.stereotype.Component
import org.springframework.web.filter.OncePerRequestFilter


@Component
class JwtAuthenticationFilter(private val jwtService: JwtService, private val userService: UserService): OncePerRequestFilter() {

    override fun doFilterInternal(
        request: HttpServletRequest,
        response: HttpServletResponse,
        filterChain: FilterChain
    ) {
        // try to get JWT in cookie or in Authorization Header
        var jwt: String? = jwtService.getJwtFromCookies(request)
        val authHeader = request.getHeader("Authorization")

        if ((jwt == null && (authHeader == null || !authHeader.startsWith("Bearer "))) || request.requestURI
                .contains("/auth")
        ) {
            filterChain.doFilter(request, response)
            return
        }

        // If the JWT is not in the cookies but in the "Authorization" header
        if (jwt == null && authHeader!!.startsWith("Bearer ")) {
            jwt = authHeader.substring(7) // after "Bearer "
        }

        val userEmail: String = jwtService.extractUserName(jwt)

        /*
           SecurityContextHolder: is where Spring Security stores the details of who is authenticated.
           Spring Security uses that information for authorization.*/
        if (StringUtils.isNotEmpty(userEmail)
            && SecurityContextHolder.getContext().authentication == null
        ) {
            val userDetails: UserDetails = userService.loadUserByUsername(userEmail)
            if (jwtService.isTokenValid(jwt, userDetails)) {
                //update the spring security context by adding a new UsernamePasswordAuthenticationToken
                val context = SecurityContextHolder.createEmptyContext()
                val authToken = UsernamePasswordAuthenticationToken(
                    userDetails,
                    null,
                    userDetails.authorities
                )
                authToken.details = WebAuthenticationDetailsSource().buildDetails(request)
                context.authentication = authToken
                SecurityContextHolder.setContext(context)
            }
        }
        filterChain.doFilter(request, response)
    }
}