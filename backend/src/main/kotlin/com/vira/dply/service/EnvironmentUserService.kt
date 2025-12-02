package com.vira.dply.service

import com.vira.dply.model.EnvironmentUser
import com.vira.dply.repository.EnvironmentUserRepository
import org.springframework.stereotype.Service

@Service
class EnvironmentUserService(private val environmentUserRepository: EnvironmentUserRepository) {

    fun findAll(): List<EnvironmentUser> {
        return environmentUserRepository.findAll()
    }

    fun findByUserId(userId: Long): List<EnvironmentUser> {
        return environmentUserRepository.findByUserId(userId)
    }

    fun findById(id: Long): EnvironmentUser? {
        return environmentUserRepository.findById(id).orElse(null)
    }

    fun create(payload: EnvironmentUser): EnvironmentUser {
        return environmentUserRepository.save(payload)
    }
}