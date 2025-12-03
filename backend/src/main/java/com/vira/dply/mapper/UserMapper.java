package com.vira.dply.mapper;

import org.mapstruct.Mapper;

import com.vira.dply.dto.UserDto;
import com.vira.dply.entity.UserEntity;

@Mapper(componentModel = "spring")
public interface UserMapper {

    UserDto toDto(UserEntity entity);

    UserEntity toEntity(UserDto dto);
}