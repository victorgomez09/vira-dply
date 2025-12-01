package com.vira.dply.service

import com.vira.dply.model.User
import com.vira.dply.repository.UserRepository
import org.springframework.security.core.userdetails.UserDetails
import org.springframework.security.core.userdetails.UserDetailsService
import org.springframework.stereotype.Service

@Service
class UserService(val userRepository: UserRepository) : UserDetailsService {

    fun findByEmail(email: String): User {
        return userRepository.findByEmail(email)
    }

    fun findAll(): List<User> {
        return userRepository.findAll()
    }

    override fun loadUserByUsername(username: String): UserDetails {
        return userRepository.findByEmail(username)
    }
}