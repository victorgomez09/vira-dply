package com.vira.dply.util;

import java.io.BufferedReader;
import java.io.File;
import java.io.InputStreamReader;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.stream.Collectors;

import org.springframework.scheduling.annotation.Async;
import org.springframework.stereotype.Component;

import com.vira.dply.entity.EnvironmentEntity;
import com.vira.dply.entity.TeamEntity;
import com.vira.dply.entity.UserTeamRoleEntity;
import com.vira.dply.enums.TeamRole;
import com.vira.dply.repository.EnvironmentRepository;

import io.fabric8.kubernetes.api.model.Namespace;
import io.fabric8.kubernetes.api.model.NamespaceBuilder;
import io.fabric8.kubernetes.api.model.rbac.RoleBinding;
import io.fabric8.kubernetes.api.model.rbac.RoleBindingBuilder;
import io.fabric8.kubernetes.client.Config;
import io.fabric8.kubernetes.client.KubernetesClient;
import io.fabric8.kubernetes.client.KubernetesClientBuilder;
import lombok.RequiredArgsConstructor;

@Component
@RequiredArgsConstructor
public class KubernetesNamespaceUtil {

    private final EnvironmentRepository environmentRepository;

    @Async
    public void createNamespaceForTeam(TeamEntity team) {
        EnvironmentEntity env = team.getEnvironment();

        try {
            String kubeConfigPath = env.getKubeConfigPath();
            String kubeConfigContent;

            if (kubeConfigPath == null || kubeConfigPath.isBlank()) {
                // Obtener kubeconfig del cluster existente vía k3d
                String clusterName = "k3d-" + env.getName(); // Ajusta según convención
                Process process = new ProcessBuilder("k3d", "kubeconfig", "get", clusterName)
                        .redirectErrorStream(true)
                        .start();

                kubeConfigContent = new BufferedReader(new InputStreamReader(process.getInputStream()))
                        .lines()
                        .collect(Collectors.joining("\n"));

                process.waitFor();

                if (kubeConfigContent.isBlank()) {
                    throw new IllegalStateException("No se pudo obtener kubeconfig del cluster existente");
                }

                // Guardar temporalmente o en BD para futuros usos
                File tmpKubeConfig = new File(
                        System.getProperty("user.home"),
                        ".paas/kubeconfigs/" + clusterName + "-kubeconfig.yaml");
                tmpKubeConfig.getParentFile().mkdirs();
                Files.writeString(tmpKubeConfig.toPath(), kubeConfigContent);
                kubeConfigPath = tmpKubeConfig.getAbsolutePath();
                env.setKubeConfigPath(kubeConfigPath);
                environmentRepository.save(env);

            } else {
                // Leer contenido desde la ruta existente
                kubeConfigContent = Files.readString(Path.of(kubeConfigPath));
            }

            // Nombre del contexto a usar (igual al cluster de k3d)
            String contextName = "k3d-" + env.getName();

            // Inicializar KubernetesClient con kubeconfig y contexto
            Config config = Config.fromKubeconfig(contextName, kubeConfigContent, kubeConfigPath);
            try (KubernetesClient client = new KubernetesClientBuilder()
                    .withConfig(config)
                    .build()) {

                // Crear Namespace para el Team
                String namespaceName = sanitizeNamespaceName(team.getName());
                Namespace ns = new NamespaceBuilder()
                        .withNewMetadata()
                        .withName(namespaceName)
                        .endMetadata()
                        .build();
                client.namespaces().createOrReplace(ns);

                // Crear RoleBindings para cada usuario del Team
                for (UserTeamRoleEntity utr : team.getUserRoles()) {
                    String roleName = mapTeamRoleToK8sRole(utr.getRole());
                    RoleBinding rb = new RoleBindingBuilder()
                            .withNewMetadata()
                            .withName(utr.getUser().getId().toString() + "-binding")
                            .withNamespace(namespaceName)
                            .endMetadata()
                            .addNewSubject()
                            .withKind("User")
                            .withName(utr.getUser().getId().toString())
                            .withApiGroup("rbac.authorization.k8s.io")
                            .endSubject()
                            .withNewRoleRef()
                            .withKind("Role")
                            .withName(roleName)
                            .withApiGroup("rbac.authorization.k8s.io")
                            .endRoleRef()
                            .build();
                    client.rbac().roleBindings().inNamespace(namespaceName).createOrReplace(rb);
                }
            }

        } catch (Exception e) {
            e.printStackTrace();
            // Opcional: marcar Team como FAILED en BD
        }
    }

        @Async
    public void syncUserRoleBinding(UserTeamRoleEntity utr) {
        TeamEntity team = utr.getTeam();
        EnvironmentEntity env = team.getEnvironment();

        try {
            String kubeConfigPath = env.getKubeConfigPath();
            if (kubeConfigPath == null || kubeConfigPath.isBlank()) {
                throw new IllegalStateException("Environment kubeConfigPath is null");
            }

            String contextName = "k3d-" + env.getName();
            Config config = Config.fromKubeconfig(contextName, Files.readString(Path.of(kubeConfigPath)), kubeConfigPath);

            try (KubernetesClient client = new KubernetesClientBuilder()
                    .withConfig(config)
                    .build()) {

                String namespaceName = sanitizeNamespaceName(team.getName());
                String roleName = mapTeamRoleToK8sRole(utr.getRole());
                String bindingName = utr.getUser().getId().toString() + "-binding";

                RoleBinding existing = client.rbac().roleBindings()
                        .inNamespace(namespaceName)
                        .withName(bindingName)
                        .get();

                if (utr.getRole() == null) {
                    // eliminar si rol es null (usuario removido)
                    if (existing != null) {
                        client.rbac().roleBindings().inNamespace(namespaceName).withName(bindingName).delete();
                    }
                    return;
                }

                // crear o actualizar
                RoleBinding rb = new RoleBindingBuilder()
                        .withNewMetadata()
                            .withName(bindingName)
                            .withNamespace(namespaceName)
                        .endMetadata()
                        .addNewSubject()
                            .withKind("User")
                            .withName(utr.getUser().getId().toString())
                            .withApiGroup("rbac.authorization.k8s.io")
                        .endSubject()
                        .withNewRoleRef()
                            .withKind("Role")
                            .withName(roleName)
                            .withApiGroup("rbac.authorization.k8s.io")
                        .endRoleRef()
                        .build();

                client.rbac().roleBindings().inNamespace(namespaceName).createOrReplace(rb);
            }

        } catch (Exception e) {
            e.printStackTrace();
            // opcional: marcar Team o Environment como FAILED en BD
        }
    }

    private String sanitizeNamespaceName(String name) {
        return name.toLowerCase().replaceAll("[^a-z0-9-]", "-");
    }

    private String mapTeamRoleToK8sRole(TeamRole role) {
        return switch (role) {
            case OWNER, ADMIN -> "admin";
            case MEMBER -> "edit";
        };
    }
}