package com.vira.dply.controller;

import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import com.vira.dply.dto.AuthDto;
import com.vira.dply.service.AuthService;

import lombok.RequiredArgsConstructor;

@RestController
@RequestMapping("/auth")
@RequiredArgsConstructor
public class AuthController {

    private final AuthService authService;

    @PostMapping("/register")
    public ResponseEntity<String> register(@RequestBody AuthDto payload) {
        return ResponseEntity.status(HttpStatus.CREATED).body(authService.register(payload.getEmail(), payload.getPassword()));
    }

    @PostMapping("/login")
    public ResponseEntity<String> login(@RequestBody AuthDto payload) {
        return ResponseEntity.status(HttpStatus.CREATED).body(authService.login(payload.getEmail(), payload.getPassword()));
    }
}