package com.vira.dply.dto;

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
    // private Set<TeamEntity> teams = new HashSet<>();
}
