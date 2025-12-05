package com.vira.dply.dto;

import java.time.Instant;
import java.util.UUID;

public record ProjectDto(
    UUID id,
    String name,
    String description,
    UUID environmentId,
    Instant createdAt
) {}