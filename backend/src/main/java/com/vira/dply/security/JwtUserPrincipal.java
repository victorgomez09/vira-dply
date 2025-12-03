package com.vira.dply.security;

import java.util.UUID;

public record JwtUserPrincipal(
    UUID userId,
    String email
) {}