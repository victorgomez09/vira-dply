package com.vira.dply.repository

import com.vira.dply.model.Environment
import org.springframework.data.jpa.repository.JpaRepository
import org.springframework.stereotype.Repository

@Repository
interface EnvironmentRepository: JpaRepository<Environment, Long>