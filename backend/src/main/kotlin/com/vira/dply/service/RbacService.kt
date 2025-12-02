package com.vira.dply.service

import com.vira.dply.model.Environment
import com.vira.dply.util.KubeconfigSecretStore
import io.kubernetes.client.openapi.ApiClient
import io.kubernetes.client.openapi.apis.CoreV1Api
import io.kubernetes.client.openapi.apis.RbacAuthorizationV1Api
import io.kubernetes.client.openapi.models.V1ObjectMeta
import io.kubernetes.client.openapi.models.V1PolicyRule
import io.kubernetes.client.openapi.models.V1Role
import io.kubernetes.client.openapi.models.V1RoleBinding
import io.kubernetes.client.openapi.models.V1RoleRef
import io.kubernetes.client.openapi.models.V1ServiceAccount
import io.kubernetes.client.openapi.models.V1Subject
import io.kubernetes.client.util.Config
import io.kubernetes.client.util.KubeConfig
import org.slf4j.LoggerFactory
import org.springframework.stereotype.Service
import java.io.StringReader

@Service
class RbacService(
    private val kubeconfigSecretStore: KubeconfigSecretStore
) {
    private val logger = LoggerFactory.getLogger(javaClass)

    fun setupTeamRbac(
        environment: Environment,
        namespace: String
    ) {
        val kubeconfig = kubeconfigSecretStore.load(environment.kubeconfigRef!!)
        val client = k8sClient(kubeconfig)

        val core = CoreV1Api(client)
        val rbac = RbacAuthorizationV1Api(client)

        createServiceAccount(core, namespace)
        createAdminRole(rbac, namespace)
        createRoleBinding(rbac, namespace)

        logger.info("RBAC configured for namespace=$namespace env=${environment.id}")
    }

    private fun k8sClient(kubeconfig: String): ApiClient {
        val kc = KubeConfig.loadKubeConfig(StringReader(kubeconfig))
        return Config.fromConfig(kc)
    }

    private fun createServiceAccount(api: CoreV1Api, namespace: String) {
        val sa = V1ServiceAccount().apply {
            metadata = V1ObjectMeta()
                .name("team-sa")
        }
        api.createNamespacedServiceAccount(namespace, sa, null, null, null, null)
    }

    private fun createAdminRole(api: RbacAuthorizationV1Api, namespace: String) {
        val role = V1Role().apply {
            metadata = V1ObjectMeta().name("team-admin")
            rules = listOf(
                V1PolicyRule()
                    .apiGroups(listOf("*"))
                    .resources(listOf("*"))
                    .verbs(listOf("*"))
            )
        }
        api.createNamespacedRole(namespace, role, null, null, null, null)
    }

    private fun createRoleBinding(api: RbacAuthorizationV1Api, namespace: String) {
        val binding = V1RoleBinding().apply {
            metadata = V1ObjectMeta().name("team-admin-binding")
            roleRef = V1RoleRef()
                .kind("Role")
                .name("team-admin")
                .apiGroup("rbac.authorization.k8s.io")

            subjects = listOf(
                V1Subject()
                    .kind("ServiceAccount")
                    .name("team-sa")
                    .namespace(namespace)
            )
        }
        api.createNamespacedRoleBinding(namespace, binding, null, null, null, null)
    }
}