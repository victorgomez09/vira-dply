package com.vira.dply.util;

import java.io.File;
import java.io.IOException;
import java.nio.file.Files;

import org.springframework.stereotype.Component;

import com.vira.dply.entity.EnvironmentEntity;

import io.fabric8.kubernetes.client.Config;
import io.fabric8.kubernetes.client.KubernetesClient;
import io.fabric8.kubernetes.client.KubernetesClientBuilder;

@Component
public class KubernetesClientProvider {

    public KubernetesClient getClientForEnvironment(EnvironmentEntity env) {
        try {
            String kubeconfigYaml = Files.readString(new File(env.getKubeConfigPath()).toPath());
            Config config = Config.fromKubeconfig(kubeconfigYaml);

            return new KubernetesClientBuilder()
                    .withConfig(config)
                    .build();

        } catch (IOException e) {
            throw new RuntimeException("Failed to read kubeconfig for environment " + env.getName(), e);
        }
    }
}
