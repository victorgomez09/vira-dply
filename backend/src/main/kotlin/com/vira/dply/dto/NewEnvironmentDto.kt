package com.vira.dply.dto

data class NewEnvironmentDto(
    val name: String,
    val kubeconfigYaml: String? = null
)
