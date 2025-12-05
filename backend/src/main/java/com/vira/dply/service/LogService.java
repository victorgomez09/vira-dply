package com.vira.dply.service;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.util.List;

import org.springframework.stereotype.Service;

import com.vira.dply.websocket.LogsWebSocketHandler;

import io.fabric8.kubernetes.api.model.Container;
import io.fabric8.kubernetes.api.model.Pod;
import io.fabric8.kubernetes.client.KubernetesClient;
import io.fabric8.kubernetes.client.dsl.LogWatch;
import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class LogService {

    private final LogsWebSocketHandler logsWebSocketHandler;

    public void streamLogs(String deployId, String namespace, String podName, KubernetesClient client) {
        Pod pod = client.pods().inNamespace(namespace).withName(podName).get();
        if (pod == null) {
            logsWebSocketHandler.sendLog(deployId, "Pod not found: " + podName);
            return;
        }

        List<Container> containers = pod.getSpec().getContainers();

        for (Container container : containers) {
            String containerName = container.getName();
            new Thread(() -> {
                try (LogWatch watch = client.pods()
                        .inNamespace(namespace)
                        .withName(podName)
                        .inContainer(containerName)
                        .watchLog(System.out)) {

                    BufferedReader reader = new BufferedReader(new InputStreamReader(watch.getOutput()));
                    String line;
                    while (!logsWebSocketHandler.isCancelled(deployId) && (line = reader.readLine()) != null) {
                        // Diferenciamos contenedor y tipo de log
                        logsWebSocketHandler.sendLog(deployId, "[" + containerName + "] " + line);
                    }

                } catch (IOException e) {
                    logsWebSocketHandler.sendLog(deployId, "[" + containerName + "] Error: " + e.getMessage());
                }
            }).start();
        }
    }

}
