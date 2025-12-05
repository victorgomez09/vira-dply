package com.vira.dply.mapper;

import org.mapstruct.Mapper;

import com.vira.dply.dto.ProjectDto;
import com.vira.dply.entity.ProjectEntity;

@Mapper(componentModel = "spring")
public interface ProjectMapper {

    ProjectEntity toEntity(ProjectDto dto);

    ProjectDto toDto(ProjectEntity entity);
}
