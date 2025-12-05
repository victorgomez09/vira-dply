package com.vira.dply.service;

import java.util.List;
import java.util.UUID;

import org.springframework.stereotype.Service;

import com.vira.dply.entity.ApplicationEntity;
import com.vira.dply.entity.ProjectEntity;
import com.vira.dply.repository.ApplicationRepository;
import com.vira.dply.repository.ProjectRepository;

import jakarta.persistence.EntityNotFoundException;
import jakarta.transaction.Transactional;
import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class ApplicationService {

        private final ApplicationRepository applicationRepository;
        private final ProjectRepository projectRepository;
        private final AccessControlService accessControlService;

        public ApplicationEntity findById(UUID projectId, UUID userId, UUID applicationId) {
                ProjectEntity project = projectRepository.findById(projectId)
                                .orElseThrow(() -> new EntityNotFoundException("Project not found"));

                accessControlService.checkUserHasAccessToEnvironment(
                                userId,
                                project.getEnvironment().getId());

                return applicationRepository.findById(applicationId)
                                .orElseThrow(() -> new EntityNotFoundException("Application not found"));
        }

        public List<ApplicationEntity> findByProjectyId(UUID projectId, UUID userId) {
                ProjectEntity project = projectRepository.findById(projectId)
                                .orElseThrow(() -> new EntityNotFoundException("Project not found"));

                accessControlService.checkUserHasAccessToEnvironment(
                                userId,
                                project.getEnvironment().getId());

                return applicationRepository.findByProjectId(projectId);
        }

        @Transactional
        public ApplicationEntity create(
                        UUID projectId,
                        UUID userId,
                        ApplicationEntity application) {
                ProjectEntity project = projectRepository.findById(projectId)
                                .orElseThrow(() -> new EntityNotFoundException("Project not found"));

                accessControlService.checkUserHasAccessToEnvironment(
                                userId,
                                project.getEnvironment().getId());

                application.setProject(project);

                return applicationRepository.save(application);
        }
}