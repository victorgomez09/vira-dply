package com.vira.dply.service;

import java.io.IOException;
import java.util.List;
import java.util.UUID;

import org.springframework.stereotype.Service;

import com.vira.dply.entity.EnvironmentEntity;
import com.vira.dply.repository.EnvironmentRepository;
import com.vira.dply.util.K3dClusterManagerUtil;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class EnvironmentService {

    private final EnvironmentRepository repository;
    private final K3dClusterManagerUtil k3dClusterManagerUtil;

    public EnvironmentEntity createEnvironment(EnvironmentEntity env) {
        if (repository.existsByName(env.getName())) {
            throw new IllegalArgumentException("Environment already exists");
        }

        if (env.getKubeConfigPath() == null || env.getKubeConfigPath().isBlank()) {
            // Crear cluster k3d por defecto y obtener kubeconfig
            String clusterName = "cluster-" + UUID.randomUUID();
            try {
                env.setKubeConfigPath(k3dClusterManagerUtil.createClusterWithDefaultConfig(clusterName));
            } catch (IOException e) {
                e.printStackTrace();
            }
        }

        env.setKubeContext(env.getName());

        return repository.save(env);
    }

    public List<EnvironmentEntity> listEnvironments() {
        return repository.findAll();
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