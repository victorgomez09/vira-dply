package com.vira.dply.dto;

import java.util.UUID;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.RequiredArgsConstructor;

@Data
@AllArgsConstructor
@RequiredArgsConstructor
public class TeamDto {
    private UUID id;
    private String name;
    private EnvironmentDto environment;
}
