package com.vira.dply.controller

import com.vira.dply.dto.LoginDto
import com.vira.dply.dto.RegisterDto
import com.vira.dply.dto.UserDto
import com.vira.dply.mapper.UserMapper
import com.vira.dply.service.AuthService
import org.springframework.http.HttpStatus
import org.springframework.http.ResponseEntity
import org.springframework.web.bind.annotation.PostMapping
import org.springframework.web.bind.annotation.RequestBody
import org.springframework.web.bind.annotation.RequestMapping
import org.springframework.web.bind.annotation.RestController

@RestController
@RequestMapping("/auth")
class AuthController(val authService: AuthService, val userMapper: UserMapper) {

    @PostMapping("/login")
    fun login(@RequestBody payload: LoginDto): ResponseEntity<String> = ResponseEntity.ok().body(authService.login(payload.email, payload.password))

    @PostMapping("/register")
    fun register(@RequestBody payload: RegisterDto): ResponseEntity<UserDto> = ResponseEntity.status(HttpStatus.CREATED).body(userMapper.toDto(authService.register(userMapper.toEntity(payload))))
}