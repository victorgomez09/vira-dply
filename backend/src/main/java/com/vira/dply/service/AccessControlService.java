package com.vira.dply.service;

import java.util.UUID;

import org.springframework.security.access.AccessDeniedException;
import org.springframework.stereotype.Service;

import com.vira.dply.repository.UserTeamRoleRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class AccessControlService {

    private final UserTeamRoleRepository userTeamRoleRepository;

    public void checkUserHasAccessToEnvironment(UUID userId, UUID environmentId) {
        boolean hasAccess = userTeamRoleRepository
                .existsByUserIdAndTeam_Environment_Id(userId, environmentId);

        if (!hasAccess) {
            throw new AccessDeniedException("No access to environment");
        }
    }
}