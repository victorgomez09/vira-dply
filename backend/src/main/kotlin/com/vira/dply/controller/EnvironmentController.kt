package com.vira.dply.controller

import com.vira.dply.dto.EnvironmentDto
import com.vira.dply.dto.NewEnvironmentDto
import com.vira.dply.mapper.EnvironmentMapper
import com.vira.dply.model.Environment
import com.vira.dply.service.EnvironmentService
import org.springframework.http.HttpStatus
import org.springframework.http.ResponseEntity
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

    @PostMapping
    fun create(@RequestBody payload: NewEnvironmentDto): ResponseEntity<EnvironmentDto> =
        ResponseEntity.status(HttpStatus.CREATED)
            .body(environmentMapper.toDto(environmentService.create(environmentMapper.toEntity(payload))))
}