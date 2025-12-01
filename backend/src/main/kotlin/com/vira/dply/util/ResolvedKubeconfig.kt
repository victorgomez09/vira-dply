package com.vira.dply.util

import io.kubernetes.client.openapi.ApiClient

data class ResolvedKubeconfig(
    val source: KubeconfigSource,
    val apiClient: ApiClient
)