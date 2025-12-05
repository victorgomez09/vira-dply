package com.vira.dply.service;

import com.vira.dply.enums.BuildStatus;
import com.vira.dply.enums.WebsocketLogType;
import com.vira.dply.util.BuildResult;
import com.vira.dply.websocket.LogsWebSocketHandler;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.time.Duration;
import java.util.UUID;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeUnit;

@Slf4j
@Service
@RequiredArgsConstructor
public class BuildService {

    private final ExecutorService executor = Executors.newCachedThreadPool();
    private final LogsWebSocketHandler logsWebSocketHandler;

    public CompletableFuture<BuildResult> buildApplicationAsync(
            String deployId,
            String repoPath,
            String imageName,
            Duration timeout
    ) {
        BuildResult result = new BuildResult();
        result.setStatus(BuildStatus.PENDING);

        return CompletableFuture.supplyAsync(() -> {
            result.setStatus(BuildStatus.BUILDING);

            ProcessBuilder pb = new ProcessBuilder();
            pb.directory(new java.io.File(repoPath));

            String tag = imageName + ":" + UUID.randomUUID().toString().substring(0, 8);
            pb.command("nixpacks", "build", ".", "--name", tag);

            try {
                Process process = pb.start();

                BufferedReader stdout = new BufferedReader(new InputStreamReader(process.getInputStream()));
                BufferedReader stderr = new BufferedReader(new InputStreamReader(process.getErrorStream()));

                ExecutorService logExecutor = Executors.newFixedThreadPool(2);

                // stdout
                logExecutor.submit(() -> stdout.lines().forEach(line -> {
                    logsWebSocketHandler.sendLog(deployId, "[BUILD] " + line, WebsocketLogType.DEPLOY);
                    log.info("[BUILD] {}", line);
                }));

                // stderr
                logExecutor.submit(() -> stderr.lines().forEach(line -> {
                    logsWebSocketHandler.sendLog(deployId, "[BUILD-ERR] " + line, WebsocketLogType.DEPLOY);
                    log.error("[BUILD] {}", line);
                }));

                boolean finished = process.waitFor(timeout.toMillis(), TimeUnit.MILLISECONDS);

                if (!finished) {
                    process.destroyForcibly();
                    result.setStatus(BuildStatus.CANCELLED);
                    logsWebSocketHandler.sendLog(deployId, "Build cancelled due to timeout", WebsocketLogType.DEPLOY);
                } else if (process.exitValue() == 0) {
                    result.setStatus(BuildStatus.SUCCESS);
                } else {
                    result.setStatus(BuildStatus.FAILED);
                }

                result.setImageTag(tag);

                logExecutor.shutdownNow();

            } catch (Exception e) {
                result.setStatus(BuildStatus.FAILED);
                logsWebSocketHandler.sendLog(deployId, "Build failed: " + e.getMessage(), WebsocketLogType.DEPLOY);
            }

            return result;
        }, executor);
    }
}
