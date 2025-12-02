package com.vira.dply.service

import com.vira.dply.model.User
import com.vira.dply.model.UserRole
import com.vira.dply.repository.UserRepository
import org.springframework.security.authentication.AuthenticationManager
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken
import org.springframework.security.core.authority.SimpleGrantedAuthority
import org.springframework.security.crypto.password.PasswordEncoder
import org.springframework.stereotype.Service


@Service
class AuthService(
    val userRepository: UserRepository,
    val passwordEncoder: PasswordEncoder,
    val authenticationManager: AuthenticationManager,
    val jwtService: JwtService,
) {

    fun register(user: User): User {
        user.encodedPassword = passwordEncoder.encode(user.password).toString()
        user.roles = listOf(UserRole.ROLE_USER)

        return userRepository.save(user)
    }

    fun login(email: String, password: String): String {
        authenticationManager.authenticate(
            UsernamePasswordAuthenticationToken(email, password)
        )

        val user: User = userRepository.findByEmail(email)

        return jwtService.generateToken(user)
    }
}