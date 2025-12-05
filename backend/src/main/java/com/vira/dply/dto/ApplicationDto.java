package com.vira.dply.dto;

import java.util.UUID;

import com.vira.dply.enums.ApplicationStatus;
import com.vira.dply.enums.ApplicationType;

public record ApplicationDto(
        UUID id,
        String name,
        String description,
        ApplicationType type,
        ApplicationStatus status,
        String gitRepository,
        String imageName) {
}