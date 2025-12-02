package com.vira.dply.mapper

import com.vira.dply.dto.TeamDto
import com.vira.dply.model.Team
import org.mapstruct.Mapper

@Mapper(componentModel = "spring")
interface TeamMapper {

    fun toDto(entity: Team): TeamDto
}