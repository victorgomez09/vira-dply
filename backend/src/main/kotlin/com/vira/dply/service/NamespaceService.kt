package com.vira.dply.service

import com.vira.dply.model.Environment
import com.vira.dply.util.KubeconfigSecretStore
import io.kubernetes.client.openapi.ApiClient
import io.kubernetes.client.openapi.apis.CoreV1Api
import io.kubernetes.client.openapi.models.V1Namespace
import io.kubernetes.client.openapi.models.V1ObjectMeta
import io.kubernetes.client.util.Config
import io.kubernetes.client.util.KubeConfig
import org.slf4j.LoggerFactory
import org.springframework.stereotype.Service
import java.io.StringReader

@Service
class NamespaceService(
    private val kubeconfigSecretStore: KubeconfigSecretStore
) {
    private val logger = LoggerFactory.getLogger(javaClass)

    fun createNamespace(environment: Environment, namespace: String) {
        val kubeconfig = kubeconfigSecretStore.load(environment.kubeconfigRef!!)
        val client = kubernetesClient(kubeconfig)

        val api = CoreV1Api(client)

        val ns = V1Namespace().apply {
            metadata = V1ObjectMeta().name(namespace)
        }

        api.createNamespace(ns, null, null, null, null)
        logger.info("Namespace $namespace created in env ${environment.id}")
    }

    private fun kubernetesClient(kubeconfig: String): ApiClient {
        val kc = KubeConfig.loadKubeConfig(StringReader(kubeconfig))
        return Config.fromConfig(kc)
    }
}