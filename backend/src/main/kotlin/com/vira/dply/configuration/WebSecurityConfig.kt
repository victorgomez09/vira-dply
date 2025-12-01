package com.vira.dply.configuration

import com.vira.dply.service.UserService
import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration
import org.springframework.security.authentication.AuthenticationManager
import org.springframework.security.authentication.AuthenticationProvider
import org.springframework.security.authentication.ProviderManager
import org.springframework.security.authentication.dao.DaoAuthenticationProvider
import org.springframework.security.config.Customizer
import org.springframework.security.config.annotation.web.builders.HttpSecurity
import org.springframework.security.config.annotation.web.configuration.EnableWebSecurity
import org.springframework.security.config.http.SessionCreationPolicy
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder
import org.springframework.security.oauth2.jwt.JwtDecoder
import org.springframework.security.oauth2.jwt.NimbusJwtDecoder
import org.springframework.security.web.SecurityFilterChain

@Configuration
@EnableWebSecurity
class WebSecurityConfig(val jwtConfig: JwtConfig, val userService: UserService) {

    @Bean
    fun securityFilterChain(http: HttpSecurity): SecurityFilterChain {
        return http
            .csrf {csrf -> csrf.disable()}
            .authorizeHttpRequests { authorizeHttpRequests -> authorizeHttpRequests.requestMatchers("/auth/**").permitAll()
                .anyRequest().authenticated() }
            .sessionManagement { sessionManagement -> sessionManagement.sessionCreationPolicy(SessionCreationPolicy.STATELESS) }
            .oauth2ResourceServer { configurer -> configurer.jwt(Customizer.withDefaults()) }
            .build()
    }

    @Bean
    fun jwtDecoder(): JwtDecoder {
        return NimbusJwtDecoder.withSecretKey(jwtConfig.getSecretKey()).build()
    }

    @Bean
    fun passwordEncoder(): BCryptPasswordEncoder {
        return BCryptPasswordEncoder()
    }

    @Bean
    fun authenticationProvider(): AuthenticationProvider {
        val authProvider: DaoAuthenticationProvider = DaoAuthenticationProvider(userService)
        authProvider.setPasswordEncoder(passwordEncoder())

        return authProvider
    }

    @Bean
    fun authManager(authenticationProvider: AuthenticationProvider): AuthenticationManager {
        return ProviderManager(authenticationProvider)
    }
}