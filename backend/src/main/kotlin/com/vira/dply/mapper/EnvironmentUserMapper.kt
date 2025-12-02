package com.vira.dply.mapper

import com.vira.dply.dto.EnvironmentUserDto
import com.vira.dply.model.EnvironmentUser
import org.mapstruct.Mapper

@Mapper(componentModel = "spring")
interface EnvironmentUserMapper {

    fun toDto(entity: EnvironmentUser): EnvironmentUserDto

    fun toEntity(dto: EnvironmentUserDto): EnvironmentUser
}