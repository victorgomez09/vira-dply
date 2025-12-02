package com.vira.dply.util

interface KubeconfigSecretStore {
    fun store(key: String, kubeconfigYaml: String): KubeconfigRef
    fun load(ref: String): String
    fun delete(ref: KubeconfigRef)
}