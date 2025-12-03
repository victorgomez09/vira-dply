package com.vira.dply.service;

import java.util.List;
import java.util.UUID;

import org.springframework.stereotype.Service;

import com.vira.dply.entity.UserTeamRoleEntity;
import com.vira.dply.repository.TeamRepository;
import com.vira.dply.repository.UserTeamRoleRepository;

import com.vira.dply.entity.TeamEntity;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class UserTeamRoleService {

    private final UserTeamRoleRepository userTeamRoleRepository;
    private final TeamRepository teamRepository;

    public List<UserTeamRoleEntity> getRolesByTeam(UUID teamUuid) {
        TeamEntity team = teamRepository.findById(teamUuid).orElseThrow(() -> new IllegalArgumentException("Team not exists in this environment"));

        return userTeamRoleRepository.findByTeam(team);
    }
}
