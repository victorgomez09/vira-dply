package com.vira.dply.controller

import com.vira.dply.dto.EnvironmentDto
import com.vira.dply.dto.NewEnvironmentDto
import com.vira.dply.mapper.EnvironmentMapper
import com.vira.dply.model.EnvironmentStatus
import com.vira.dply.service.EnvironmentService
import org.springframework.http.HttpStatus
import org.springframework.http.ResponseEntity
import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.PathVariable
import org.springframework.web.bind.annotation.PostMapping
import org.springframework.web.bind.annotation.RequestBody
import org.springframework.web.bind.annotation.RequestMapping
import org.springframework.web.bind.annotation.RestController

@RestController
@RequestMapping("/environments")
class EnvironmentController(
    private val environmentService: EnvironmentService,
    private val environmentMapper: EnvironmentMapper
) {

    @GetMapping("/{id}")
    suspend fun getEnvironment(@PathVariable id: Long): ResponseEntity<EnvironmentDto> {
        val env = environmentService.findById(id)
            ?: return ResponseEntity.notFound().build()

        return ResponseEntity.ok().body(environmentMapper.toDto(env))
    }

    @PostMapping
    fun create(@RequestBody payload: NewEnvironmentDto): ResponseEntity<EnvironmentDto> =
        ResponseEntity.status(HttpStatus.CREATED)
            .body(environmentMapper.toDto(environmentService.create(environmentMapper.toEntity(payload))))

    @PostMapping("/{id}/cancel")
    suspend fun cancelProvision(@PathVariable id: Long): ResponseEntity<String> {
        val env = environmentService.findById(id)
            ?: return ResponseEntity.notFound().build()

        if (env.status != EnvironmentStatus.PROVISIONING) {
            return ResponseEntity.status(HttpStatus.BAD_REQUEST)
                .body(env.id.toString() + env.status.name)
        }

        environmentService.cancelProvision(id)
        return ResponseEntity.ok(env.id.toString() + "CANCELLED")
    }
}