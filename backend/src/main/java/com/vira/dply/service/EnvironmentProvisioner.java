package com.vira.dply.service;

import java.util.UUID;

import org.springframework.scheduling.annotation.Async;
import org.springframework.stereotype.Service;

import com.vira.dply.entity.EnvironmentEntity;
import com.vira.dply.enums.EnvironmentStatus;
import com.vira.dply.repository.EnvironmentRepository;
import com.vira.dply.util.K3dClusterManagerUtil;

import jakarta.transaction.Transactional;
import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class EnvironmentProvisioner {

    private final EnvironmentRepository repository;
    private final K3dClusterManagerUtil k3dService;

    @Async
    @Transactional
    public void provisionAsync(UUID environmentId) {
        EnvironmentEntity env = repository.findById(environmentId).orElse(null);
        if (env == null) return;

        try {
            env.setStatus(EnvironmentStatus.CREATING);
            repository.save(env);

            // K3dClusterResult result = k3dService.createCluster(env.getName());
            k3dService.createCluster(env.getName());

            // env.setKubeConfigPath(result.kubeConfigPath());
            env.setStatus(EnvironmentStatus.READY);

            repository.save(env);
        } catch (Exception e) {
            env.setStatus(EnvironmentStatus.FAILED);
            repository.save(env);
        }
    }
}