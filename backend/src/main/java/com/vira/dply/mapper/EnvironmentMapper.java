package com.vira.dply.mapper;

import org.mapstruct.Mapper;

import com.vira.dply.dto.EnvironmentDto;
import com.vira.dply.entity.EnvironmentEntity;

@Mapper(componentModel = "spring")
public interface EnvironmentMapper {

    EnvironmentDto toDto(EnvironmentEntity entity);

    EnvironmentEntity toEntity(EnvironmentDto dto);
}
