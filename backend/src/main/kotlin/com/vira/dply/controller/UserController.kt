package com.vira.dply.controller

import org.springframework.web.bind.annotation.RestController
import org.springframework.web.bind.annotation.GetMapping

@RestController("/users")
class UserController {

    @GetMapping("/")
    fun findAll(): String = ""
}
