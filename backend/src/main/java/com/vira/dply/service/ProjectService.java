package com.vira.dply.service;

import java.util.List;
import java.util.UUID;

import org.springframework.stereotype.Service;

import com.vira.dply.entity.EnvironmentEntity;
import com.vira.dply.entity.ProjectEntity;
import com.vira.dply.repository.EnvironmentRepository;
import com.vira.dply.repository.ProjectRepository;

import jakarta.persistence.EntityNotFoundException;
import jakarta.transaction.Transactional;
import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class ProjectService {

    private final ProjectRepository projectRepository;
    private final EnvironmentRepository environmentRepository;
    private final AccessControlService accessControlService;

    public List<ProjectEntity> findByEnvironment(UUID envId, UUID userId) {
        accessControlService.checkUserHasAccessToEnvironment(userId, envId);

        return projectRepository.findByEnvironmentId(envId);
    }

    @Transactional
    public ProjectEntity createProject(
            UUID envId,
            UUID userId,
            ProjectEntity payload
    ) {
        accessControlService.checkUserHasAccessToEnvironment(userId, envId);

        EnvironmentEntity env = environmentRepository.findById(envId)
                .orElseThrow(() -> new EntityNotFoundException("Environment not found"));

        payload.setEnvironment(env);

        return projectRepository.save(payload);
    }

    public ProjectEntity findById(UUID projectId, UUID userId) {
        ProjectEntity project = projectRepository.findById(projectId)
                .orElseThrow(() -> new EntityNotFoundException("Project not found"));

        accessControlService.checkUserHasAccessToEnvironment(
                userId,
                project.getEnvironment().getId()
        );

        return project;
    }
}