package com.vira.dply.dto

data class EnvironmentUserDto(
    val id: Long,
    val environment: EnvironmentDto,
    val user: UserDto,
    val roles: List<String>
)
