package com.vira.dply.mapper

import com.vira.dply.dto.RegisterDto
import com.vira.dply.dto.UserDto
import com.vira.dply.model.User
import org.mapstruct.Mapper

@Mapper(componentModel = "spring")
interface UserMapper {

    fun toEntity(dto: RegisterDto): User

    fun toDto(entity: User): UserDto
}