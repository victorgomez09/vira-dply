package com.vira.dply.dto

import java.util.Date

data class ExceptionDto(
    val timestamp: Date,
    val status: Int,
    val error: String,
    val message: String,
    val path: String
)
