package com.vira.dply.mapper

import com.vira.dply.dto.EnvironmentDto
import com.vira.dply.dto.NewEnvironmentDto
import com.vira.dply.model.Environment
import org.mapstruct.Mapper

@Mapper(componentModel = "spring")
interface EnvironmentMapper {

    fun toEntity(dto: NewEnvironmentDto): Environment

    fun toDto(entity: Environment): EnvironmentDto
}