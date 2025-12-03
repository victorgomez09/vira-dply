package com.vira.dply.service;

import java.util.List;
import java.util.UUID;

import org.springframework.stereotype.Service;

import com.vira.dply.entity.EnvironmentEntity;
import com.vira.dply.entity.TeamEntity;
import com.vira.dply.entity.UserTeamRoleEntity;
import com.vira.dply.enums.TeamRole;
import com.vira.dply.repository.EnvironmentRepository;
import com.vira.dply.repository.TeamRepository;
import com.vira.dply.repository.UserRepository;
import com.vira.dply.repository.UserTeamRoleRepository;
import com.vira.dply.util.K3dClusterManagerUtil;
import com.vira.dply.util.StringUtil;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class EnvironmentService {

    private final EnvironmentRepository repository;
    private final K3dClusterManagerUtil k3dClusterManagerUtil;
    private final EnvironmentProvisioner environmentProvisioner;
    private final TeamRepository teamRepository;
    private final UserTeamRoleRepository userTeamRoleRepository;
    private final UserRepository userRepository;
    private final StringUtil stringUtil;

    public EnvironmentEntity createEnvironment(EnvironmentEntity env, UUID creator) {
        if (repository.existsByName(env.getName())) {
            throw new IllegalArgumentException("Environment already exists");
        }

        String clusterName = stringUtil.removeTrailingDash(("cluster-" + UUID.randomUUID()).substring(0, 32));
        env.setKubeContext(clusterName);

        env = repository.save(env);

        environmentProvisioner.provisionAsync(env.getId());

        TeamEntity team = new TeamEntity();
        team.setName("default");
        team.setEnvironment(env);

        team = teamRepository.save(team);

        UserTeamRoleEntity utr = new UserTeamRoleEntity();
        utr.setUser(userRepository.findById(creator).orElseThrow(() -> new IllegalArgumentException()));
        utr.setTeam(team);
        utr.setRole(TeamRole.OWNER);

        userTeamRoleRepository.save(utr);

        return env;
    }

    public List<EnvironmentEntity> listEnvironments(UUID id) {
        return repository.findAllAccessibleByUser(id);
    }

    public EnvironmentEntity getEnvironment(UUID id) {
        return repository.findById(id)
                .orElseThrow(() -> new IllegalArgumentException("Environment not found"));
    }

    public void deleteEnvironment(UUID id) {
        EnvironmentEntity env = repository.findById(id)
                .orElseThrow(() -> new IllegalArgumentException("Environment not found"));

        k3dClusterManagerUtil.deleteCluster(env.getName());

        repository.delete(env);
    }
}