package com.vira.dply.mapper;

import org.mapstruct.Mapper;

import com.vira.dply.dto.TeamDto;
import com.vira.dply.entity.TeamEntity;

@Mapper(componentModel = "spring")
public interface TeamMapper {

    TeamDto toDto(TeamEntity entity);

    TeamEntity toEntity(TeamDto dto);
}
