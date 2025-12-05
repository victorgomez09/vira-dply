package com.vira.dply.service;

import java.util.Collections;

import org.springframework.stereotype.Service;

import com.vira.dply.entity.ApplicationEntity;

import io.fabric8.kubernetes.api.model.IntOrString;
import io.fabric8.kubernetes.api.model.ServiceBuilder;
import io.fabric8.kubernetes.api.model.apps.Deployment;
import io.fabric8.kubernetes.api.model.apps.DeploymentBuilder;
import io.fabric8.kubernetes.client.KubernetesClient;
import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class DeploymentService {

    private final KubernetesClient kubernetesClient;

    public void deployApplication(ApplicationEntity app, String imageTag) {
        // Ejemplo: Deployment y Service
        String namespace = app.getProject().getEnvironment().getName();

        // Crear Deployment
        Deployment deployment = new DeploymentBuilder()
                .withNewMetadata().withName(app.getName()).withNamespace(namespace).endMetadata()
                .withNewSpec()
                    .withReplicas(1)
                    .withNewSelector()
                        .addToMatchLabels("app", app.getName())
                    .endSelector()
                    .withNewTemplate()
                        .withNewMetadata().addToLabels("app", app.getName()).endMetadata()
                        .withNewSpec()
                            .addNewContainer()
                                .withName(app.getName())
                                .withImage(imageTag)
                                .addNewPort().withContainerPort(8080).endPort()
                            .endContainer()
                        .endSpec()
                    .endTemplate()
                .endSpec()
                .build();

        kubernetesClient.apps().deployments().inNamespace(namespace).createOrReplace(deployment);

        // Crear Service
        io.fabric8.kubernetes.api.model.Service service = new ServiceBuilder()
                .withNewMetadata().withName(app.getName()).withNamespace(namespace).endMetadata()
                .withNewSpec()
                    .addNewPort().withPort(80).withTargetPort(new IntOrString(8080)).endPort()
                    .withSelector(Collections.singletonMap("app", app.getName()))
                    .withType("ClusterIP")
                .endSpec()
                .build();

        kubernetesClient.services().inNamespace(namespace).createOrReplace(service);
    }
}
