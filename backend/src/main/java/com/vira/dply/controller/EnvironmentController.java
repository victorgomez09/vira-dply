package com.vira.dply.controller;

import java.util.List;
import java.util.UUID;

import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import com.vira.dply.dto.EnvironmentDto;
import com.vira.dply.entity.EnvironmentEntity;
import com.vira.dply.mapper.EnvironmentMapper;
import com.vira.dply.service.EnvironmentService;

import lombok.RequiredArgsConstructor;

@RestController
@RequestMapping("/api/environments")
@RequiredArgsConstructor
public class EnvironmentController {

    private final EnvironmentService service;
    private final EnvironmentMapper environmentMapper;

    @PostMapping
    public ResponseEntity<EnvironmentDto> create(@RequestBody EnvironmentDto payload) {
        return ResponseEntity.status(HttpStatus.CREATED).body(environmentMapper.toDto(service.createEnvironment(environmentMapper.toEntity(payload))));
    }

    @GetMapping
    public List<EnvironmentEntity> list() {
        return service.listEnvironments();
    }

    @GetMapping("/{id}")
    public EnvironmentEntity get(@PathVariable UUID id) {
        return service.getEnvironment(id);
    }

    @DeleteMapping("/{id}")
    public void delete(@PathVariable UUID id) {
        service.deleteEnvironment(id);
    }
}