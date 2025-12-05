package com.vira.dply.mapper;

import org.mapstruct.Mapper;

import com.vira.dply.dto.ApplicationDto;
import com.vira.dply.entity.ApplicationEntity;

@Mapper(componentModel = "spring")
public interface ApplicationMapper {

    ApplicationEntity toEntity(ApplicationDto dto);

    ApplicationDto toDto(ApplicationEntity entity);
}
