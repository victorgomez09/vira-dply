package com.vira.dply.service;

import java.util.List;
import java.util.UUID;

import org.springframework.stereotype.Service;

import com.vira.dply.entity.EnvironmentEntity;
import com.vira.dply.entity.TeamEntity;
import com.vira.dply.entity.UserTeamRoleEntity;
import com.vira.dply.enums.TeamRole;
import com.vira.dply.repository.TeamRepository;
import com.vira.dply.repository.UserRepository;
import com.vira.dply.repository.UserTeamRoleRepository;
import com.vira.dply.util.KubernetesClientProvider;
import com.vira.dply.util.KubernetesNamespaceUtil;

import io.fabric8.kubernetes.client.KubernetesClient;
import jakarta.transaction.Transactional;
import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class TeamService {

    private final TeamRepository teamRepository;
    private final EnvironmentService environmentService;
    private final UserRepository userRepository;
    private final UserTeamRoleRepository userTeamRoleRepository;
    private final KubernetesNamespaceUtil kubernetesNamespaceUtil;
    private final KubernetesClientProvider kubernetesClientProvider;

    @Transactional
    public TeamEntity createTeam(String name, UUID environmentId, UUID userId) {
        EnvironmentEntity env = environmentService.getEnvironment(environmentId);

        if (teamRepository.existsByNameAndEnvironment_Id(name, environmentId)) {
            throw new IllegalArgumentException("Team already exists in this environment");
        }

        // Guardar en DB
        TeamEntity team = new TeamEntity();
        team.setName(name);
        team.setEnvironment(env);
        team = teamRepository.save(team);

        kubernetesNamespaceUtil.createNamespaceForTeam(team);

        UserTeamRoleEntity utr = new UserTeamRoleEntity();
        utr.setUser(userRepository.findById(userId).orElseThrow(() -> new IllegalArgumentException()));
        utr.setTeam(team);
        utr.setRole(TeamRole.OWNER);
        userTeamRoleRepository.save(utr);

        return team;
    }

    @Transactional
    public TeamEntity addUserToTeam(UUID teamId, UUID userId, TeamRole role) {
        TeamEntity team = teamRepository.findById(teamId)
                .orElseThrow(() -> new IllegalArgumentException("Team not found"));

        UserTeamRoleEntity utr = new UserTeamRoleEntity();
        utr.setTeam(team);
        utr.setUser(userRepository.findById(userId).orElseThrow(() -> new IllegalArgumentException()));
        utr.setRole(role);

        utr = userTeamRoleRepository.save(utr);

        kubernetesNamespaceUtil.syncUserRoleBinding(utr);

        return team;
    }

    @Transactional
    public TeamEntity updateUserRole(UUID teamId, UUID userId, TeamRole newRole) {
        UserTeamRoleEntity utr = userTeamRoleRepository.findByTeamIdAndUserId(teamId, userId)
                .orElseThrow(() -> new IllegalArgumentException("User not in team"));

        utr.setRole(newRole);
        utr = userTeamRoleRepository.save(utr);

        kubernetesNamespaceUtil.syncUserRoleBinding(utr);

        return teamRepository.findById(teamId).orElseThrow(() -> new IllegalArgumentException());
    }

    @Transactional
    public void removeUserFromTeam(UUID teamId, UUID userId) {
        UserTeamRoleEntity utr = userTeamRoleRepository.findByTeamIdAndUserId(teamId, userId)
                .orElseThrow(() -> new IllegalArgumentException("User not in team"));

        kubernetesNamespaceUtil.syncUserRoleBinding(utr);

        userTeamRoleRepository.delete(utr);
    }

    public List<TeamEntity> listTeams(UUID environmentId) {
        EnvironmentEntity env = environmentService.getEnvironment(environmentId);
        return teamRepository.findAll().stream()
                .filter(t -> t.getEnvironment().equals(env))
                .toList();
    }

    public void deleteTeam(UUID teamId) {
        TeamEntity team = teamRepository.findById(teamId)
                .orElseThrow(() -> new IllegalArgumentException("Team not found"));

        EnvironmentEntity env = team.getEnvironment();

        try (KubernetesClient client = kubernetesClientProvider.getClientForEnvironment(env)) {
            client.namespaces().withName(team.getName()).delete();
        } catch (Exception e) {
            throw new RuntimeException("Failed to delete namespace in cluster " + env.getName(), e);
        }

        teamRepository.delete(team);
    }
}