package com.vira.dply.git.dto

data class GitSource(
    val url: String,
    val ref: String = "main"
)