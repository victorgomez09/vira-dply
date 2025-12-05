package com.vira.dply.service;

import java.nio.file.Path;
import java.time.Duration;
import java.util.UUID;
import java.util.concurrent.CompletableFuture;

import org.springframework.scheduling.annotation.Async;
import org.springframework.stereotype.Service;

import com.vira.dply.entity.ApplicationEntity;
import com.vira.dply.enums.BuildStatus;
import com.vira.dply.enums.WebsocketLogType;
import com.vira.dply.repository.ApplicationRepository;
import com.vira.dply.util.BuildResult;
import com.vira.dply.websocket.LogsWebSocketHandler;

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
    private final LogsWebSocketHandler logsWebSocketHandler;

    @Async
    @Transactional
    public void buildAndDeployApplication(UUID applicationId) {
        ApplicationEntity app = applicationRepository.findById(applicationId)
                .orElseThrow(() -> new EntityNotFoundException("Application not found"));

        String deployId = UUID.randomUUID().toString();

        try {
            // 1️⃣ Clonar repo con streaming de logs
            logsWebSocketHandler.sendLog(deployId, "Starting Git clone...", WebsocketLogType.DEPLOY);
            Path repoPath = gitService.cloneRepo(deployId, app.getGitRepository(), app.getGitBranch(), null);
            logsWebSocketHandler.sendLog(deployId, "Git clone completed.", WebsocketLogType.DEPLOY);

            // 2️⃣ Actualizar estado
            app.setBuildStatus(BuildStatus.BUILDING);
            applicationRepository.save(app);

            // 3️⃣ Ejecutar build con streaming de logs
            logsWebSocketHandler.sendLog(deployId, "Starting build...", WebsocketLogType.DEPLOY);
            CompletableFuture<BuildResult> buildFuture = buildService.buildApplicationAsync(
                    deployId,
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
                    logsWebSocketHandler.sendLog(deployId, "Build successful. Deploying application...", WebsocketLogType.DEPLOY);
                    deploymentService.deployApplication(deployId, app, result.getImageTag());
                } else {
                    logsWebSocketHandler.sendLog(deployId, "Build failed. Deployment skipped.", WebsocketLogType.DEPLOY);
                }
            }).exceptionally(ex -> {
                app.setBuildStatus(BuildStatus.FAILED);
                app.setBuildLogs(ex.getMessage());
                applicationRepository.save(app);
                logsWebSocketHandler.sendLog(deployId, "Build failed: " + ex.getMessage(), WebsocketLogType.DEPLOY);
                return null;
            });

        } catch (Exception e) {
            app.setBuildStatus(BuildStatus.FAILED);
            app.setBuildLogs(e.getMessage());
            applicationRepository.save(app);
            logsWebSocketHandler.sendLog(deployId, "Error during deployment process: " + e.getMessage(), WebsocketLogType.DEPLOY);
        }
    }
}
