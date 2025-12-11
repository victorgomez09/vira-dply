package com.vira.dply.dto;

import java.util.List;
import java.util.UUID;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@AllArgsConstructor
@NoArgsConstructor
public class EnvironmentDto {
    private UUID id;
    private String name;
    private String kubeContext;
    private String kubeConfigPath;
    private List<TeamDto> teams;
}
