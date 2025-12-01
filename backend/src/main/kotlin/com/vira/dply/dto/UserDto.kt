package com.vira.dply.dto

import java.util.Date

data class UserDto(
    val id: Long,
    val fullName: String,
    val email: String,
    val role: String,
    val createdAt: Date,
    val updatedDate: Date
)