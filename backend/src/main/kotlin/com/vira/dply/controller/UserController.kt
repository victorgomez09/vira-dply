package com.vira.dply.controller

import org.springframework.web.bind.annotation.RestController
import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.RequestMapping

@RestController
@RequestMapping("/users")
class UserController {

    @GetMapping("/")
    fun findAll(): String = ""
}
