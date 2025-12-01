package com.vira.dply.util

data class KubeconfigRef(
    val type: String,  // "file"
    val id: String     // nombre del fichero
)