package com.vira.dply.service

import com.vira.dply.model.Environment
import com.vira.dply.model.EnvironmentStatus
import com.vira.dply.repository.EnvironmentRepository
import com.vira.dply.util.KubeconfigSecretStore
import com.vira.dply.util.KubernetesClientUtil
import io.kubernetes.client.openapi.ApiClient
import io.kubernetes.client.openapi.Configuration
import io.kubernetes.client.openapi.apis.CoreV1Api
import io.kubernetes.client.util.Config
import io.kubernetes.client.util.KubeConfig
import jakarta.transaction.Transactional
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.launch
import org.slf4j.LoggerFactory
import org.springframework.stereotype.Service
import java.io.StringReader
import java.time.Instant
import java.util.concurrent.TimeUnit

@Service
class EnvironmentService(
    private val environmentRepository: EnvironmentRepository,
    private val kubeconfigSecretStore: KubeconfigSecretStore
) {
    private val logger = LoggerFactory.getLogger(javaClass)

    // Scope dedicado para provisión de clusters
    private val provisionScope = CoroutineScope(SupervisorJob() + Dispatchers.IO)

    fun create(payload: Environment): Environment {
        // 1️⃣ Crear Environment en BD
        val savedEnv = environmentRepository.save(payload)

        // 2️⃣ Lanzar provisión asíncrona
        provisionScope.launch {
            try {
                provisionCluster(savedEnv)
            } catch (ex: Exception) {
                logger.error("Error provisioning cluster ${savedEnv.id}", ex)
                savedEnv.status = EnvironmentStatus.FAILED
                environmentRepository.save(savedEnv)
            }
        }

        return savedEnv
    }

    private suspend fun provisionCluster(env: Environment) {
        val clusterName = "env-${env.id}"

        // 1) Crear cluster k3d
        runProcess(listOf("k3d", "cluster", "create", clusterName, "--wait"))

        // 2) Obtener kubeconfig
        val kubeconfig = runProcessCaptureOutput(listOf("k3d", "kubeconfig", "get", clusterName))

        // 3) Validar cluster con client-java
        val client = createKubernetesClient(kubeconfig)
        validateClusterReady(client)

        // 4) Guardar kubeconfig en secret store
        val kubeRef = kubeconfigSecretStore.store(env.id.toString(), kubeconfig)

        // 5) Actualizar Environment
        env.kubeconfigRef = kubeRef.toString()
        env.status = EnvironmentStatus.READY
        environmentRepository.save(env)
    }

    private fun createKubernetesClient(kubeconfigYaml: String): ApiClient {
        val kubeConfig = KubeConfig.loadKubeConfig(StringReader(kubeconfigYaml))
        val client = Config.fromConfig(kubeConfig)
        io.kubernetes.client.openapi.Configuration.setDefaultApiClient(client)
        return client
    }

    private fun validateClusterReady(client: ApiClient) {
        val coreV1 = CoreV1Api(client)
        val nodes = coreV1.listNode(null, null, null, null, null, null, null, null, 5, false)
        if (nodes.items.isEmpty()) throw IllegalStateException("No nodes found in cluster")
    }

    private fun runProcess(command: List<String>) {
        val process = ProcessBuilder(command)
            .redirectErrorStream(true)
            .start()
        val ok = process.waitFor(5, TimeUnit.MINUTES)
        val output = process.inputStream.bufferedReader().readText()
        if (!ok || process.exitValue() != 0) {
            throw IllegalStateException("Command failed: ${command.joinToString(" ")}\n$output")
        }
    }

    private fun runProcessCaptureOutput(command: List<String>): String {
        val process = ProcessBuilder(command)
            .redirectErrorStream(true)
            .start()
        val ok = process.waitFor(2, TimeUnit.MINUTES)
        val output = process.inputStream.bufferedReader().readText()
        if (!ok || process.exitValue() != 0) {
            throw IllegalStateException("Command failed: ${command.joinToString(" ")}\n$output")
        }
        return output
    }
}