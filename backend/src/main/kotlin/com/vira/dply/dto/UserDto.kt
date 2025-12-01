package com.vira.dply.dto

import java.util.Date

data class UserDto(
    var id: Long,
    var fullName: String,
    var email: String,
    var role: String,
    var createdAt: Date,
    var updatedDate: Date
)