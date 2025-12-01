package com.vira.dply.dto

import java.time.Instant

data class EnvironmentDto(
    val id: Long,
    val name: String,
    var type: String,
    var status: String,
    var kubeconfigRef: String?,
    val createdAt: Instant,
    val updatedAt: Instant
)