package com.vira.dply.service;

import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.UUID;
import java.util.concurrent.CompletableFuture;

import org.springframework.stereotype.Service;

import com.vira.dply.entity.ApplicationEntity;
import com.vira.dply.entity.DomainEntity;
import com.vira.dply.enums.GatewayType;
import com.vira.dply.util.KubernetesClientProvider;
import com.vira.dply.websocket.LogsWebSocketHandler;

import io.fabric8.kubernetes.api.model.GenericKubernetesResource;
import io.fabric8.kubernetes.api.model.IntOrString;
import io.fabric8.kubernetes.api.model.ObjectMetaBuilder;
import io.fabric8.kubernetes.api.model.Pod;
import io.fabric8.kubernetes.api.model.ServiceBuilder;
import io.fabric8.kubernetes.api.model.apps.Deployment;
import io.fabric8.kubernetes.api.model.apps.DeploymentBuilder;
import io.fabric8.kubernetes.client.KubernetesClient;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

@Slf4j
@Service
@RequiredArgsConstructor
public class DeploymentService {

        private final KubernetesClientProvider kubernetesClientProvider;
        private final LogsWebSocketHandler logsWebSocketHandler;
        private final LogService logService;

        public String deployApplication(ApplicationEntity app, String imageTag) {
                String deployId = UUID.randomUUID().toString();

                // Ejecutamos deploy de forma asíncrona
                CompletableFuture.runAsync(() -> {
                        try {
                                logsWebSocketHandler.sendLog(deployId, "Starting deployment of " + app.getName());

                                String namespace = app.getProject().getEnvironment().getName();
                                Set<DomainEntity> domains = app.getDomains();

                                boolean hasDomains = domains != null && !domains.isEmpty();
                                Integer containerPort = hasDomains ? new ArrayList<>(domains).get(0).getPort() : null;

                                // 1️⃣ Crear Deployment
                                DeploymentBuilder deploymentBuilder = new DeploymentBuilder()
                                                .withNewMetadata().withName(app.getName()).withNamespace(namespace)
                                                .endMetadata()
                                                .withNewSpec()
                                                .withReplicas(1)
                                                .withNewSelector().addToMatchLabels("app", app.getName()).endSelector()
                                                .withNewTemplate()
                                                .withNewMetadata().addToLabels("app", app.getName()).endMetadata()
                                                .withNewSpec()
                                                .addNewContainer()
                                                .withName(app.getName())
                                                .withImage(imageTag)
                                                .addNewPort().withContainerPort(containerPort).endPort() // <-- solo
                                                                                                         // aquí
                                                .endContainer()
                                                .endSpec()
                                                .endTemplate()
                                                .endSpec();

                                if (hasDomains) {
                                        deploymentBuilder.editSpec().editTemplate().editSpec()
                                                        .editContainer(0)
                                                        .addNewPort().withContainerPort(containerPort).endPort()
                                                        .endContainer()
                                                        .endSpec()
                                                        .endTemplate().endSpec();
                                }

                                Deployment deployment = deploymentBuilder.build();
                                KubernetesClient client = kubernetesClientProvider
                                                .getClientForEnvironment(app.getProject().getEnvironment());
                                client.apps().deployments().inNamespace(namespace).createOrReplace(deployment);
                                logsWebSocketHandler.sendLog(deployId, "Deployment created.");

                                log.info("Deployment created for app {}", app.getName());

                                // 2️⃣ Si tiene dominios, crear Service y HTTPRoutes
                                if (hasDomains) {
                                        createService(app, namespace, containerPort, client);
                                        createHttpRoutes(app, namespace, domains, client);
                                }

                                String podName = waitForPodReady(namespace, app.getName(), client);
                                logsWebSocketHandler.sendLog(deployId, "Streaming logs from pod: " + podName);
                                logService.streamLogs(deployId, namespace, podName, client);

                                logsWebSocketHandler.sendLog(deployId, "Deployment completed.");
                        } catch (Exception e) {
                                logsWebSocketHandler.sendLog(deployId, "Error: " + e.getMessage());
                                e.printStackTrace();
                        }
                });

                return deployId;
        }

        private void createService(ApplicationEntity app, String namespace, int port, KubernetesClient client) {
                io.fabric8.kubernetes.api.model.Service service = new ServiceBuilder()
                                .withNewMetadata().withName(app.getName()).withNamespace(namespace).endMetadata()
                                .withNewSpec()
                                .addNewPort().withPort(port).withTargetPort(new IntOrString(port)).endPort()
                                .withSelector(Collections.singletonMap("app", app.getName()))
                                .withType("ClusterIP")
                                .endSpec()
                                .build();

                client.services().inNamespace(namespace).createOrReplace(service);
                log.info("Service created for app {}", app.getName());
        }

        private void createHttpRoutes(ApplicationEntity app, String namespace, Set<DomainEntity> domains,
                        KubernetesClient client) {
                for (DomainEntity domain : domains) {

                        createGatewayIfNotExists(namespace, domain, client);

                        // TLS
                        if (Boolean.TRUE.equals(domain.getTls())) {
                                createCertificate(app, namespace, domain, client);
                        }

                        Map<String, Object> spec = new HashMap<>();
                        spec.put("parentRefs", List
                                        .of(Map.of("name", domain.getGatewayType().name().toLowerCase() + "-gateway")));
                        spec.put("hostnames", List.of(domain.getHost(), "localhost")); // localhost opcional
                        spec.put("rules", List.of(Map.of(
                                        "matches", List.of(Map.of("path", Map.of("type", "Prefix", "value", "/"))),
                                        "forwardTo",
                                        List.of(Map.of("serviceName", app.getName(), "port", domain.getPort())))));

                        if (Boolean.TRUE.equals(domain.getTls())) {
                                spec.put("tls", Map.of("secretName",
                                                app.getName() + "-tls-" + domain.getHost().replace(".", "-")));
                        }

                        GenericKubernetesResource httpRoute = new GenericKubernetesResource();
                        httpRoute.setApiVersion("gateway.networking.k8s.io/v1beta1");
                        httpRoute.setKind("HTTPRoute");
                        httpRoute.setMetadata(new ObjectMetaBuilder()
                                        .withName(app.getName() + "-route-" + domain.getHost().replace(".", "-"))
                                        .withNamespace(namespace)
                                        .build());
                        httpRoute.setAdditionalProperty("spec", spec);

                        client.genericKubernetesResources(
                                        "gateway.networking.k8s.io/v1beta1", "httproutes")
                                        .inNamespace(namespace)
                                        .createOrReplace(httpRoute);

                        log.info("HTTPRoute created for domain {} in namespace {} using gateway {}", domain.getHost(),
                                        namespace,
                                        domain.getGatewayType());
                }
        }

        private void createCertificate(ApplicationEntity app, String namespace, DomainEntity domain,
                        KubernetesClient client) {
                String secretName = app.getName() + "-tls-" + domain.getHost().replace(".", "-");

                GenericKubernetesResource certificate = new GenericKubernetesResource();
                certificate.setApiVersion("cert-manager.io/v1");
                certificate.setKind("Certificate");
                certificate.setMetadata(new ObjectMetaBuilder()
                                .withName(secretName)
                                .withNamespace(namespace)
                                .build());

                Map<String, Object> spec = Map.of(
                                "secretName", secretName,
                                "dnsNames", List.of(domain.getHost()),
                                "issuerRef", Map.of(
                                                "name", "letsencrypt-staging", // o letsencrypt-prod
                                                "kind", "ClusterIssuer"));

                certificate.setAdditionalProperty("spec", spec);

                client.genericKubernetesResources("cert-manager.io/v1", "certificates")
                                .inNamespace(namespace)
                                .createOrReplace(certificate);

                log.info("Certificate created for domain {} in namespace {}", domain.getHost(), namespace);
        }

        private void createGatewayIfNotExists(String namespace, DomainEntity domain, KubernetesClient client) {
                String gatewayName = domain.getGatewayType().name().toLowerCase() + "-gateway";

                // Comprobamos si ya existe
                GenericKubernetesResource existing = client.genericKubernetesResources(
                                "gateway.networking.k8s.io/v1beta1", "gateways")
                                .inNamespace(namespace)
                                .withName(gatewayName)
                                .get();

                if (existing != null)
                        return;

                GenericKubernetesResource gateway = new GenericKubernetesResource();
                gateway.setApiVersion("gateway.networking.k8s.io/v1beta1");
                gateway.setKind("Gateway");
                gateway.setMetadata(new ObjectMetaBuilder()
                                .withName(gatewayName)
                                .withNamespace(namespace)
                                .build());

                // Selección del controlador según el tipo
                Map<String, Object> spec = new HashMap<>();
                if (domain.getGatewayType() == GatewayType.TRAEFIK) {
                        spec.put("gatewayClassName", "traefik");
                } else {
                        spec.put("gatewayClassName", "nginx");
                }

                // Listener HTTP/HTTPS
                spec.put("listeners", List.of(
                                Map.of(
                                                "name", "http",
                                                "protocol", "HTTP",
                                                "port", 80),
                                Map.of(
                                                "name", "https",
                                                "protocol", "HTTPS",
                                                "port", 443)));

                gateway.setAdditionalProperty("spec", spec);

                client.genericKubernetesResources(
                                "gateway.networking.k8s.io/v1beta1", "gateways")
                                .inNamespace(namespace)
                                .createOrReplace(gateway);

                log.info("Gateway {} created in namespace {}", gatewayName, namespace);
        }

        private String waitForPodReady(String namespace, String appName, KubernetesClient client)
                        throws InterruptedException {
                // Esperar a que aparezca un pod con label app=appName
                for (int i = 0; i < 30; i++) { // timeout 30s
                        Pod pod = client.pods().inNamespace(namespace)
                                        .withLabel("app", appName)
                                        .list().getItems().stream().findFirst().orElse(null);
                        if (pod != null && pod.getStatus().getPhase().equals("Running")) {
                                return pod.getMetadata().getName();
                        }
                        Thread.sleep(1000);
                }
                throw new RuntimeException("Pod not ready after timeout");
        }

}
