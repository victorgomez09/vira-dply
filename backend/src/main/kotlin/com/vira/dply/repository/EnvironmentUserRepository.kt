package com.vira.dply.repository

import com.vira.dply.model.EnvironmentUser
import org.springframework.data.jpa.repository.JpaRepository
import org.springframework.stereotype.Repository

@Repository
interface EnvironmentUserRepository: JpaRepository<EnvironmentUser, Long> {

    fun findByUserId(userId: Long): List<EnvironmentUser>
}