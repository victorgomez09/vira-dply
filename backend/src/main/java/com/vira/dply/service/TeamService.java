package com.vira.dply.service;

import java.util.List;
import java.util.UUID;

import org.springframework.stereotype.Service;

import com.vira.dply.entity.EnvironmentEntity;
import com.vira.dply.entity.TeamEntity;
import com.vira.dply.repository.TeamRepository;
import com.vira.dply.util.KubernetesClientProvider;

import io.fabric8.kubernetes.client.KubernetesClient;
import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class TeamService {

    private final TeamRepository teamRepository;
    private final EnvironmentService environmentService;
    private final TeamRoleSyncService teamRoleSyncService;
    private final KubernetesClientProvider kubernetesClientProvider;

    public TeamEntity createTeam(String name, UUID environmentId) {
        EnvironmentEntity env = environmentService.getEnvironment(environmentId);

        if (teamRepository.existsByNameAndEnvironment_Id(name, environmentId)) {
            throw new IllegalArgumentException("Team already exists in this environment");
        }

        try (KubernetesClient client = kubernetesClientProvider.getClientForEnvironment(env)) {
            client.namespaces()
                    .createOrReplace(
                            new io.fabric8.kubernetes.api.model.NamespaceBuilder()
                                    .withNewMetadata()
                                    .withName(name)
                                    .endMetadata()
                                    .build());

        } catch (Exception e) {
            throw new RuntimeException("Failed to create namespace in cluster " + env, e);
        }

        // Guardar en DB
        TeamEntity team = new TeamEntity();
        team.setName(name);
        team.setEnvironment(env);
        team = teamRepository.save(team);

        teamRoleSyncService.syncTeamRoles(team);

        return team;
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