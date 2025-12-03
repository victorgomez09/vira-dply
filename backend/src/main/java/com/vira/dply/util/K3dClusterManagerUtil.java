package com.vira.dply.util;

import java.io.File;
import java.io.IOException;
import java.nio.file.Files;
import java.util.List;

import org.springframework.stereotype.Component;

import lombok.RequiredArgsConstructor;

@Component
@RequiredArgsConstructor
public class K3dClusterManagerUtil {

    private final ProcessExecutorUtil processExecutorUtil;
    private final PortUtil portUtil;

    /**
     * Crea un cluster k3d con un nombre único
     * 
     * @throws IOException
     */
    public String createClusterWithDefaultConfig(String name) throws IOException {
        int apiPort = portUtil.findFreePort();

        File baseDir = new File(System.getProperty("user.home"), ".paas/kubeconfigs");
        if (!baseDir.exists()) {
            boolean created = baseDir.mkdirs();
            if (!created) {
                throw new RuntimeException("Could not create directory: " + baseDir.getAbsolutePath());
            }
        }

        String kubeConfigPath = createDefaultKubeConfig(
                name,
                apiPort);
        List<String> command = List.of(
                "k3d", "cluster", "create", name,
                "--config", kubeConfigPath, "--wait");

        processExecutorUtil.run(command);
        // executeCommand(command);

        return kubeConfigPath;
    }

    /** Crea un cluster k3d con un nombre único */
    public void createCluster(String name) {
        List<String> command = List.of(
                "k3d", "cluster", "create", name, "--wait");

        processExecutorUtil.run(command);
        // executeCommand(command);
    }

    /** Borra un cluster k3d */
    public void deleteCluster(String name) {
        List<String> command = List.of(
                "k3d", "cluster", "delete", name);

        processExecutorUtil.run(command);
        // executeCommand(command);
    }

    private String createDefaultKubeConfig(
            String envName,
            int apiPort) {
        try {
            File baseDir = new File(
                    System.getProperty("user.home"),
                    ".paas/kubeconfigs");

            if (!baseDir.exists() && !baseDir.mkdirs()) {
                throw new RuntimeException(
                        "Cannot create kubeconfig directory: " + baseDir.getAbsolutePath());
            }

            File kubeConfig = new File(
                    baseDir,
                    envName + "-kubeconfig.yaml");

            String kubeconfigYaml = """
                    apiVersion: v1
                    kind: Config

                    clusters:
                      - name: default
                        cluster:
                          server: https://127.0.0.1:%d
                          insecure-skip-tls-verify: true

                    contexts:
                      - name: default
                        context:
                          cluster: default
                          user: default

                    current-context: default

                    users:
                      - name: default
                        user:
                          token: dummy-token
                    """.formatted(apiPort);

            Files.writeString(kubeConfig.toPath(), kubeconfigYaml);

            // ✅ Validación fuerte
            String written = Files.readString(kubeConfig.toPath());
            if (!written.contains("apiVersion: v1")
                    || !written.contains("127.0.0.1:" + apiPort)) {
                throw new RuntimeException("Invalid kubeconfig generated");
            }

            return kubeConfig.getAbsolutePath();
        } catch (Exception e) {
            throw new RuntimeException(
                    "Failed to create default kubeconfig", e);
        }
    }
}