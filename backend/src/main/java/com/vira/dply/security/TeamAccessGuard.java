package com.vira.dply.security;

import java.util.Set;
import java.util.UUID;

import org.springframework.stereotype.Component;

import com.vira.dply.enums.Role;
import com.vira.dply.repository.UserTeamRoleRepository;

import lombok.RequiredArgsConstructor;

@Component
@RequiredArgsConstructor
public class TeamAccessGuard {

    private final UserTeamRoleRepository repo;

    public void check(UUID userId, UUID teamId, Set<Role> allowedRoles) {
        boolean allowed = repo.existsByUser_IdAndTeam_IdAndRoleIn(
                userId,
                teamId,
                allowedRoles);

        if (!allowed) {
            throw new SecurityException("Forbidden for team");
        }
    }
}