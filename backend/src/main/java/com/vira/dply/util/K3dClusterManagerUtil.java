package com.vira.dply.util;

import java.io.File;
import java.io.IOException;
import java.util.List;

import org.springframework.stereotype.Component;

import lombok.RequiredArgsConstructor;

@Component
@RequiredArgsConstructor
public class K3dClusterManagerUtil {

    private final ProcessExecutorUtil processExecutorUtil;

    /** Crea un cluster k3d con un nombre único 
     * @throws IOException */
    public String createClusterWithDefaultConfig(String name) throws IOException {
        File tempKubeConfig = File.createTempFile(name + "-kubeconfig", ".yaml");
        List<String> command = List.of(
                "k3d", "cluster", "create", name,
                    "--kubeconfig", tempKubeConfig.getAbsolutePath(), "--wait");

        processExecutorUtil.run(command);
        // executeCommand(command);

        return tempKubeConfig.getAbsolutePath();
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
}