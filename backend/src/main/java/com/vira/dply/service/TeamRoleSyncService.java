package com.vira.dply.service;

import java.util.List;

import org.springframework.stereotype.Service;

import com.vira.dply.entity.TeamEntity;
import com.vira.dply.entity.UserTeamRoleEntity;
import com.vira.dply.util.KubeRoleMapper;
import com.vira.dply.util.KubernetesClientProvider;

import io.fabric8.kubernetes.api.model.ObjectMetaBuilder;
import io.fabric8.kubernetes.api.model.rbac.RoleBinding;
import io.fabric8.kubernetes.api.model.rbac.RoleRef;
import io.fabric8.kubernetes.api.model.rbac.Subject;
import io.fabric8.kubernetes.client.KubernetesClient;
import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class TeamRoleSyncService {

    private final UserTeamRoleService userTeamRoleService;
    private final KubernetesClientProvider kubernetesClientProvider;
    private final KubeRoleMapper kubeRoleMapper;

    /**
     * Sincroniza los RoleBindings del team en Kubernetes
     */
    public void syncTeamRoles(TeamEntity team) {
        try (KubernetesClient client = kubernetesClientProvider.getClientForEnvironment(team.getEnvironment())) {

            for (UserTeamRoleEntity utr : userTeamRoleService.getRolesByTeam(team.getId())) {
                Subject subject = new Subject();
                subject.setKind("User");
                subject.setName(utr.getUser().getEmail());
                subject.setNamespace(team.getName());

                // crear RoleBinding
                RoleBinding rb = new RoleBinding();
                rb.setMetadata(new ObjectMetaBuilder()
                        .withName("user-" + utr.getUser().getId() + "-binding")
                        .withNamespace(team.getName())
                        .build());

                RoleRef roleRef = new RoleRef();
                roleRef.setApiGroup("rbac.authorization.k8s.io");
                roleRef.setKind("ClusterRole");
                roleRef.setName(kubeRoleMapper.toKubeRole(utr.getRole()));
                rb.setRoleRef(roleRef);

                rb.setSubjects(List.of(subject));

                client.rbac().roleBindings()
                        .inNamespace(team.getName())
                        .createOrReplace(rb);
            }
        }
    }
}