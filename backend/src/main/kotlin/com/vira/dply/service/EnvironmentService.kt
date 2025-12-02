package com.vira.dply.service

import com.vira.dply.model.Environment
import com.vira.dply.model.EnvironmentRole
import com.vira.dply.model.EnvironmentStatus
import com.vira.dply.model.EnvironmentUser
import com.vira.dply.model.User
import com.vira.dply.repository.EnvironmentRepository
import com.vira.dply.repository.EnvironmentUserRepository
import com.vira.dply.util.KubeconfigSecretStore
import io.kubernetes.client.openapi.ApiClient
import io.kubernetes.client.openapi.Configuration
import io.kubernetes.client.openapi.apis.CoreV1Api
import io.kubernetes.client.util.Config
import io.kubernetes.client.util.KubeConfig
import kotlinx.coroutines.*
import org.slf4j.LoggerFactory
import org.springframework.stereotype.Service
import java.io.StringReader
import java.util.concurrent.TimeUnit
import kotlin.math.min
import kotlin.random.Random

@Service
class EnvironmentService(
    private val environmentRepository: EnvironmentRepository,
    private val environmentUserRepository: EnvironmentUserRepository,
    private val kubeconfigSecretStore: KubeconfigSecretStore
) {
    private val logger = LoggerFactory.getLogger(javaClass)
    // Scope dedicado para todas las provisiones
    private val provisionScope = CoroutineScope(SupervisorJob() + Dispatchers.IO)
    // Map para manejar cancelación de provisiones
    private val cancellationMap = mutableMapOf<Long, Job>()

    fun findAll(): List<Environment> {
        return environmentRepository.findAll()
    }

    fun findById(id: Long): Environment? {
        return environmentRepository.findById(id).orElse(null)
    }

    fun create(payload: Environment, user: User): Environment {
        val savedEnv = environmentRepository.save(payload)
        val environmetUser = EnvironmentUser(
            id = 0L,
            environment = savedEnv,
            user = null,
            role = EnvironmentRole.ADMIN
        )

        val job = provisionScope.launch {
            try {
                provisionClusterWithRetry(savedEnv)
            } catch (ex: CancellationException) {
                logger.warn("Provisioning cancelled for ${savedEnv.id}")
                savedEnv.status = EnvironmentStatus.FAILED
                environmentRepository.save(savedEnv)
            } catch (ex: Exception) {
                logger.error("Error provisioning cluster ${savedEnv.id}", ex)
                savedEnv.status = EnvironmentStatus.FAILED
                environmentRepository.save(savedEnv)
            } finally {
                cancellationMap.remove(savedEnv.id)
            }
        }

        cancellationMap[savedEnv.id] = job
        return savedEnv
    }

    fun cancelProvision(envId: Long) {
        cancellationMap[envId]?.cancel()
    }

    fun delete(envId: Long) {
        val env = findById(envId) ?: return
        val clusterName = "env-${env.id}"

        try {
            logger.info("Deleting k3d cluster $clusterName")
            runProcess(listOf("k3d", "cluster", "delete", clusterName))
        } catch (ex: Exception) {
            logger.error("Error deleting cluster $clusterName", ex)
        }

        environmentRepository.deleteById(envId)
    }

    private suspend fun provisionClusterWithRetry(env: Environment) {
        val maxAttempts = 3
        var delayMs = 1000L

        repeat(maxAttempts) { attempt ->
            try {
                provisionCluster(env)
                return  // éxito, salimos
            } catch (ex: Exception) {
                if (attempt == maxAttempts - 1) throw ex  // último intento, lanza excepción
                logger.warn("Attempt ${attempt + 1} failed for cluster ${env.id}, retrying in $delayMs ms", ex)
                delay(delayMs)
                delayMs = min(delayMs * 2, 10000L) + Random.nextLong(0, 500)
            }
        }
    }

    private suspend fun provisionCluster(env: Environment) {
        val clusterName = "env-${env.id}"
        logger.info("Starting provisioning for cluster $clusterName")

        // 1) Crear cluster k3d
        logger.info("Creating k3d cluster $clusterName")
        runProcess(listOf("k3d", "cluster", "create", clusterName, "--wait"))

        // 2) Obtener kubeconfig
        logger.info("Retrieving kubeconfig for $clusterName")
        val kubeconfig = runProcessCaptureOutput(listOf("k3d", "kubeconfig", "get", clusterName))

        // 3) Validar cluster con client-java
        logger.info("Validating cluster $clusterName")
        val client = createKubernetesClient(kubeconfig)
        validateClusterReady(client)

        // 4) Guardar kubeconfig
        logger.info("Storing kubeconfig for ${env.id}")
        val kubeRef = kubeconfigSecretStore.store(env.id.toString(), kubeconfig)

        // 5) Actualizar Environment
        env.kubeconfigRef = kubeRef.toString()
        env.status = EnvironmentStatus.READY
        environmentRepository.save(env)
        logger.info("Provisioning completed for cluster $clusterName")
    }

    private fun createKubernetesClient(kubeconfigYaml: String): ApiClient {
        val kubeConfig = KubeConfig.loadKubeConfig(StringReader(kubeconfigYaml))
        val client = Config.fromConfig(kubeConfig)
        Configuration.setDefaultApiClient(client)
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
