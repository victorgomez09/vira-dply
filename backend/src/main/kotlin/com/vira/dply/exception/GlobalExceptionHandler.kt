package com.vira.dply.exception

import com.vira.dply.dto.ExceptionDto
import org.springframework.web.bind.annotation.ExceptionHandler
import org.springframework.web.bind.annotation.RestControllerAdvice
import org.springframework.web.context.request.WebRequest

@RestControllerAdvice
class GlobalExceptionHandler {

    @ExceptionHandler(Exception::class)
    fun handleAllExceptions(ex: Exception, webRequest: WebRequest): ExceptionDto {
        return ExceptionDto(
            timestamp = java.util.Date(),
            status = 500,
            error = "Internal Server Error",
            message = ex.message ?: "An unexpected error occurred",
            path = webRequest.getDescription(false).replace("uri=", "")
        )
    }
}