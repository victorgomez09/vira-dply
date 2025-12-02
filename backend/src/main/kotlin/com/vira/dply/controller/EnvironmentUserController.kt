package com.vira.dply.controller

import com.vira.dply.dto.EnvironmentUserDto
import com.vira.dply.mapper.EnvironmentUserMapper
import com.vira.dply.model.User
import com.vira.dply.service.EnvironmentUserService
import org.springframework.http.ResponseEntity
import org.springframework.security.core.annotation.AuthenticationPrincipal
import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.RequestMapping
import org.springframework.web.bind.annotation.RestController

@RestController
@RequestMapping("/environment-users")
class EnvironmentUserController(private val environmentUserService: EnvironmentUserService, private val environmentUserMapper: EnvironmentUserMapper) {

    @GetMapping
    fun findAllByUser(@AuthenticationPrincipal user: User): ResponseEntity<List<EnvironmentUserDto>> {
        return ResponseEntity.ok(environmentUserService.findByUserId(user.id).map {e -> environmentUserMapper.toDto(e) }.toList())
    }
}