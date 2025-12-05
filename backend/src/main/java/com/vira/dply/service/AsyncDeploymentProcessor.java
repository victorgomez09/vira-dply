package com.vira.dply.service;

import java.nio.file.Path;
import java.time.Duration;
import java.util.UUID;
import java.util.concurrent.CompletableFuture;

import org.springframework.scheduling.annotation.Async;
import org.springframework.stereotype.Service;

import com.vira.dply.entity.ApplicationEntity;
import com.vira.dply.enums.BuildStatus;
import com.vira.dply.repository.ApplicationRepository;
import com.vira.dply.util.BuildResult;

import jakarta.persistence.EntityNotFoundException;
import jakarta.transaction.Transactional;
import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class AsyncDeploymentProcessor {

    private final ApplicationRepository applicationRepository;
    private final GitService gitService;
    private final BuildService buildService;
    private final DeploymentService deploymentService;

    @Async
    @Transactional
    public void buildAndDeployApplication(UUID applicationId) {
        ApplicationEntity app = applicationRepository.findById(applicationId).orElseThrow(() -> new EntityNotFoundException("Application not found"));

        try {
            // 1️⃣ Clonar repo
            Path repoPath = gitService.cloneRepo(app.getGitRepository(), app.getGitBranch(), null);
            
            // 2️⃣ Actualizar estado
            app.setBuildStatus(BuildStatus.BUILDING);
            applicationRepository.save(app);

            // 3️⃣ Ejecutar build con Nixpacks
            CompletableFuture<BuildResult> buildFuture = buildService.buildApplicationAsync(
                    repoPath.toString(),
                    app.getName(),
                    Duration.ofMinutes(5)
            );

            buildFuture.thenAccept(result -> {
                app.setBuildStatus(result.getStatus());
                app.setBuildLogs(result.getLogs());
                app.setImageName(result.getImageTag());
                applicationRepository.save(app);

                // 4️⃣ Si build exitoso, desplegar en Kubernetes
                if (result.getStatus() == BuildStatus.SUCCESS) {
                    deploymentService.deployApplication(app, result.getImageTag());
                }
            }).exceptionally(ex -> {
                app.setBuildStatus(BuildStatus.FAILED);
                app.setBuildLogs(ex.getMessage());
                applicationRepository.save(app);
                return null;
            });

        } catch (Exception e) {
            app.setBuildStatus(BuildStatus.FAILED);
            app.setBuildLogs(e.getMessage());
            applicationRepository.save(app);
        }
    }
}
