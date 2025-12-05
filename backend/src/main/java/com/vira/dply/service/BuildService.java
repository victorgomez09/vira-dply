package com.vira.dply.service;

import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import com.vira.dply.enums.BuildStatus;
import com.vira.dply.util.BuildResult;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.time.Duration;
import java.util.UUID;
import java.util.concurrent.*;

@Slf4j
@Service
public class BuildService {

    private final ExecutorService executor = Executors.newCachedThreadPool();

    public CompletableFuture<BuildResult> buildApplicationAsync(String repoPath, String imageName, Duration timeout) {
        BuildResult result = new BuildResult();
        result.setStatus(BuildStatus.PENDING);

        CompletableFuture<BuildResult> future = CompletableFuture.supplyAsync(() -> {
            result.setStatus(BuildStatus.BUILDING);

            ProcessBuilder pb = new ProcessBuilder();
            pb.directory(new java.io.File(repoPath));

            // Comando Nixpacks
            String tag = imageName + ":" + UUID.randomUUID().toString().substring(0, 8);
            pb.command("nixpacks", "build", ".", "--name", tag);

            try {
                Process process = pb.start();

                // Logs en streaming
                StringBuilder logs = new StringBuilder();
                BufferedReader reader = new BufferedReader(new InputStreamReader(process.getInputStream()));
                BufferedReader errReader = new BufferedReader(new InputStreamReader(process.getErrorStream()));

                ExecutorService logExecutor = Executors.newFixedThreadPool(2);
                logExecutor.submit(() -> reader.lines().forEach(line -> {
                    logs.append(line).append("\n");
                    log.info("[NIXPACKS] {}", line);
                }));
                logExecutor.submit(() -> errReader.lines().forEach(line -> {
                    logs.append(line).append("\n");
                    log.error("[NIXPACKS] {}", line);
                }));

                boolean finished = process.waitFor(timeout.toMillis(), TimeUnit.MILLISECONDS);

                if (!finished) {
                    process.destroyForcibly();
                    result.setStatus(BuildStatus.CANCELLED);
                    logs.append("Build cancelled due to timeout\n");
                } else if (process.exitValue() == 0) {
                    result.setStatus(BuildStatus.SUCCESS);
                } else {
                    result.setStatus(BuildStatus.FAILED);
                }

                result.setLogs(logs.toString());
                result.setImageTag(tag);

                logExecutor.shutdownNow();
                return result;

            } catch (Exception e) {
                result.setStatus(BuildStatus.FAILED);
                result.setLogs(e.getMessage());
                return result;
            }

        }, executor);

        return future;
    }

    public void cancelBuild(CompletableFuture<BuildResult> buildFuture) {
        buildFuture.cancel(true);
    }
}
