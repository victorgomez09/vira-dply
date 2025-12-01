package com.vira.dply.util

import com.vira.dply.util.KubeconfigSource
import com.vira.dply.util.ResolvedKubeconfig

import io.kubernetes.client.util.Config
import io.kubernetes.client.openapi.Configuration
import io.kubernetes.client.util.KubeConfig
import org.springframework.stereotype.Component
import java.io.File
import java.io.StringReader

@Component
class KubernetesClientUtil {
    fun resolve(optionalYaml: String?): ResolvedKubeconfig {
        return when {
            !optionalYaml.isNullOrBlank() -> {
                createFromYaml(optionalYaml, KubeconfigSource.REQUEST_BODY)
            }

            System.getenv("KUBECONFIG") != null -> {
                createFromFile(
                    File(System.getenv("KUBECONFIG")),
                    KubeconfigSource.KUBECONFIG_ENV
                )
            }

            File(System.getProperty("user.home"), ".kube/config").exists() -> {
                createFromFile(
                    File(System.getProperty("user.home"), ".kube/config"),
                    KubeconfigSource.DEFAULT_FILE
                )
            }

            else -> error("No kubeconfig provided and no default found")
        }
    }

    private fun createFromYaml(
        yaml: String,
        source: KubeconfigSource
    ): ResolvedKubeconfig {

        val kubeConfig = KubeConfig.loadKubeConfig(StringReader(yaml))
        val client = Config.fromConfig(kubeConfig)

        Configuration.setDefaultApiClient(client)
        return ResolvedKubeconfig(source, client)
    }

    private fun createFromFile(
        file: File,
        source: KubeconfigSource
    ): ResolvedKubeconfig {

        val kubeConfig = KubeConfig.loadKubeConfig(file.reader())
        val client = Config.fromConfig(kubeConfig)

        Configuration.setDefaultApiClient(client)
        return ResolvedKubeconfig(source, client)
    }
}